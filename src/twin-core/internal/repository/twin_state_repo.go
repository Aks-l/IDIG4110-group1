package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"IDIG4110/shared/dto"
	"IDIG4110/twin-core/internal/db"
	"IDIG4110/twin-core/internal/domain"

	"github.com/jackc/pgx/v5"
)

const (
	// Write path, auto-provisions placeholder device and entity on first sight

	ensureDefaultHomeQuery = `
		INSERT INTO homes (id, name)
		VALUES ($1::uuid, 'Unassigned')
		ON CONFLICT (id) DO NOTHING
	`

	insertDeviceQuery = `
		INSERT INTO devices (home_id, gateway_id, external_id, name)
		VALUES ($1::uuid, $2::uuid, $3, $4)
		ON CONFLICT (gateway_id, external_id) DO NOTHING
		RETURNING id::text
	`

	getDeviceQuery = `
		SELECT id::text FROM devices
		WHERE gateway_id = $1::uuid AND external_id = $2
	`

	insertEntityQuery = `
		INSERT INTO entities (device_id, external_entity_id, name, domain, device_class, unit)
		VALUES ($1::uuid, $2, $3, $4, $5, $6)
		ON CONFLICT (device_id, external_entity_id) DO NOTHING
		RETURNING id::text
	`

	getEntityByDeviceQuery = `
		SELECT id::text FROM entities
		WHERE device_id = $1::uuid AND external_entity_id = $2
	`

	// Readings arrive with the source's entity id
	resolveEntityQuery = `
		SELECT e.id::text
		FROM entities e
		JOIN devices d ON d.id = e.device_id
		WHERE d.gateway_id = $1::uuid AND e.external_entity_id = $2
		ORDER BY e.updated_at DESC
		LIMIT 1
	`

	// Last-write-wins on reading time, superseded values move to previous_*
	upsertStateQuery = `
		INSERT INTO twin_state (entity_id, value_num, value_text, attributes, updated_at)
		VALUES ($1::uuid, $2, $3, $4, $5)
		ON CONFLICT (entity_id) DO UPDATE
		SET previous_value_num = CASE WHEN twin_state.updated_at < EXCLUDED.updated_at
			    THEN twin_state.value_num ELSE twin_state.previous_value_num END,
		    previous_value_text = CASE WHEN twin_state.updated_at < EXCLUDED.updated_at
			    THEN twin_state.value_text ELSE twin_state.previous_value_text END,
		    previous_attributes = CASE WHEN twin_state.updated_at < EXCLUDED.updated_at
			    THEN twin_state.attributes ELSE twin_state.previous_attributes END,
		    previous_updated_at = CASE WHEN twin_state.updated_at < EXCLUDED.updated_at
			    THEN twin_state.updated_at ELSE twin_state.previous_updated_at END,
		    value_num  = EXCLUDED.value_num,
		    value_text = EXCLUDED.value_text,
		    attributes = EXCLUDED.attributes,
		    updated_at = EXCLUDED.updated_at
		WHERE EXCLUDED.updated_at >= twin_state.updated_at
	`

	// Read path

	listHomesQuery = `
		SELECT id::text, name, address, timezone
		FROM homes
		ORDER BY name
	`

	getHomeQuery = `
		SELECT id::text, name, address, timezone
		FROM homes
		WHERE id = $1::uuid
	`

	listAreasByHomeQuery = `
		SELECT id::text, home_id::text, name, floor, area_type, geometry
		FROM areas
		WHERE home_id = $1::uuid
		ORDER BY name
	`

	listDevicesByHomeQuery = `
		SELECT id::text, home_id::text, area_id::text, gateway_id::text,
			external_id, name, manufacturer, model, sw_version, device_type
		FROM devices
		WHERE home_id = $1::uuid
		ORDER BY name
	`

	// uuid columns cast to text, scans into Go strings
	entityStateColumns = `
		e.id::text,
		e.external_entity_id,
		COALESCE(e.name, e.external_entity_id),
		e.domain,
		e.controllable,
		COALESCE(e.device_class, ''),
		COALESCE(e.unit, ''),
		e.area_id::text,
		d.id::text,
		d.home_id::text,
		d.gateway_id::text,
		ts.value_num,
		ts.value_text,
		ts.attributes,
		ts.updated_at,
		ts.previous_value_num,
		ts.previous_value_text,
		ts.previous_attributes,
		ts.previous_updated_at
	`

	listEntityStatesByHomeQuery = `
		SELECT` + entityStateColumns + `
		FROM entities e
		JOIN devices d ON d.id = e.device_id
		LEFT JOIN twin_state ts ON ts.entity_id = e.id
		WHERE d.home_id = $1::uuid
		ORDER BY e.external_entity_id
	`

	getEntityStateQuery = `
		SELECT` + entityStateColumns + `
		FROM entities e
		JOIN devices d ON d.id = e.device_id
		LEFT JOIN twin_state ts ON ts.entity_id = e.id
		WHERE e.id = $1::uuid
	`
)

type TwinStateRepoImpl struct {
	db *db.Database
}

func NewTwinStateRepoImpl(db *db.Database) *TwinStateRepoImpl {
	return &TwinStateRepoImpl{
		db: db,
	}
}

// Applies one reading to twin_state
// resolves entity by (gateway_id, external_entity_id), auto-provisions
// placeholder device and entity on first sight
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - reading [dto.Reading] the reading to apply
//
// # Returns:
//
//   - Error on resolve, provision or upsert failure
func (r *TwinStateRepoImpl) ApplyReading(ctx context.Context, reading dto.Reading) error {
	tx, err := r.db.Conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var entityID string
	err = tx.QueryRow(ctx, resolveEntityQuery, reading.GatewayID, reading.ExternalEntityID).Scan(&entityID)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		entityID, err = r.provisionEntity(ctx, tx, reading)
		if err != nil {
			return err
		}
	case err != nil:
		return fmt.Errorf("resolve entity: %w", err)
	}

	if _, err := tx.Exec(ctx, upsertStateQuery, entityID, reading.ValueNum, reading.ValueText, attributesParam(reading), reading.Timestamp); err != nil {
		return fmt.Errorf("upsert twin state: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// Registers placeholder device and entity for unseen entity
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - tx [pgx.Tx] transaction
//   - reading [dto.Reading] the reading being applied
//
// # Returns:
//
//   - New entity id
//   - Error on failure
func (r *TwinStateRepoImpl) provisionEntity(ctx context.Context, tx pgx.Tx, reading dto.Reading) (string, error) {
	if _, err := tx.Exec(ctx, ensureDefaultHomeQuery, domain.DefaultHomeID); err != nil {
		return "", fmt.Errorf("ensure default home: %w", err)
	}

	deviceID, err := r.ensureDevice(ctx, tx, reading)
	if err != nil {
		return "", err
	}

	entityID, err := r.ensureEntity(ctx, tx, deviceID, reading)
	if err != nil {
		return "", err
	}

	slog.Info("auto-provisioned placeholder entity",
		"gateway_id", reading.GatewayID,
		"external_entity_id", reading.ExternalEntityID,
		"entity_id", entityID,
	)
	return entityID, nil
}

func (r *TwinStateRepoImpl) ensureDevice(ctx context.Context, tx pgx.Tx, reading dto.Reading) (string, error) {
	var deviceID string
	err := tx.QueryRow(
		ctx,
		insertDeviceQuery,
		domain.DefaultHomeID,
		reading.GatewayID,
		reading.ExternalEntityID,
		reading.ExternalEntityID,
	).Scan(&deviceID)
	if errors.Is(err, pgx.ErrNoRows) {
		// Already provisioned, or lost the insert race: fetch the row.
		err = tx.QueryRow(ctx, getDeviceQuery, reading.GatewayID, reading.ExternalEntityID).Scan(&deviceID)
	}
	if err != nil {
		return "", fmt.Errorf("ensure device: %w", err)
	}
	return deviceID, nil
}

func (r *TwinStateRepoImpl) ensureEntity(ctx context.Context, tx pgx.Tx, deviceID string, reading dto.Reading) (string, error) {
	var deviceClass *string
	if reading.DeviceClass != "" {
		deviceClass = &reading.DeviceClass
	}
	var unit *string
	if reading.Unit != "" {
		unit = &reading.Unit
	}

	var entityID string
	err := tx.QueryRow(
		ctx,
		insertEntityQuery,
		deviceID,
		reading.ExternalEntityID,
		entityName(reading),
		domain.EntityDomain(reading.ExternalEntityID),
		deviceClass,
		unit,
	).Scan(&entityID)
	if errors.Is(err, pgx.ErrNoRows) {
		// Already provisioned, or lost the insert race: fetch the row.
		err = tx.QueryRow(ctx, getEntityByDeviceQuery, deviceID, reading.ExternalEntityID).Scan(&entityID)
	}
	if err != nil {
		return "", fmt.Errorf("ensure entity: %w", err)
	}
	return entityID, nil
}

func (r *TwinStateRepoImpl) ListHomes(ctx context.Context) ([]domain.HomeSummary, error) {
	rows, err := r.db.Conn.Query(ctx, listHomesQuery)
	if err != nil {
		return nil, fmt.Errorf("list homes: %w", err)
	}
	defer rows.Close()

	homes := []domain.HomeSummary{}
	for rows.Next() {
		var h domain.HomeSummary
		if err := rows.Scan(&h.ID, &h.Name, &h.Address, &h.Timezone); err != nil {
			return nil, fmt.Errorf("list homes: %w", err)
		}
		homes = append(homes, h)
	}
	return homes, rows.Err()
}

func (r *TwinStateRepoImpl) GetHome(ctx context.Context, homeID string) (domain.HomeSummary, error) {
	var h domain.HomeSummary
	err := r.db.Conn.QueryRow(ctx, getHomeQuery, homeID).Scan(&h.ID, &h.Name, &h.Address, &h.Timezone)
	if errors.Is(err, pgx.ErrNoRows) {
		return h, domain.ErrNotFound
	}
	if err != nil {
		return h, fmt.Errorf("get home: %w", err)
	}
	return h, nil
}

func (r *TwinStateRepoImpl) ListAreasByHome(ctx context.Context, homeID string) ([]domain.Area, error) {
	rows, err := r.db.Conn.Query(ctx, listAreasByHomeQuery, homeID)
	if err != nil {
		return nil, fmt.Errorf("list areas: %w", err)
	}
	defer rows.Close()

	areas := []domain.Area{}
	for rows.Next() {
		var a domain.Area
		if err := rows.Scan(&a.ID, &a.HomeID, &a.Name, &a.Floor, &a.AreaType, &a.Geometry); err != nil {
			return nil, fmt.Errorf("list areas: %w", err)
		}
		areas = append(areas, a)
	}
	return areas, rows.Err()
}

func (r *TwinStateRepoImpl) ListDevicesByHome(ctx context.Context, homeID string) ([]domain.Device, error) {
	rows, err := r.db.Conn.Query(ctx, listDevicesByHomeQuery, homeID)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	devices := []domain.Device{}
	for rows.Next() {
		var d domain.Device
		if err := rows.Scan(&d.ID, &d.HomeID, &d.AreaID, &d.GatewayID, &d.ExternalID, &d.Name, &d.Manufacturer, &d.Model, &d.SwVersion, &d.DeviceType); err != nil {
			return nil, fmt.Errorf("list devices: %w", err)
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

// Returns one page of the entity list
// filters combine with AND, zero filter lists everything
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - filter [domain.EntityFilter] list filters
//
// # Returns:
//
//   - Page of entity states
//   - Error on query failure
func (r *TwinStateRepoImpl) ListEntityStates(ctx context.Context, filter domain.EntityFilter) (domain.EntityStatePage, error) {
	conditions := []string{}
	args := []any{}

	if filter.HomeID != "" {
		args = append(args, filter.HomeID)
		conditions = append(conditions, fmt.Sprintf("d.home_id = $%d::uuid", len(args)))
	}
	if filter.AreaID != "" {
		args = append(args, filter.AreaID)
		conditions = append(conditions, fmt.Sprintf("e.area_id = $%d::uuid", len(args)))
	}
	if filter.DeviceID != "" {
		args = append(args, filter.DeviceID)
		conditions = append(conditions, fmt.Sprintf("e.device_id = $%d::uuid", len(args)))
	}
	if len(filter.Domains) > 0 {
		args = append(args, filter.Domains)
		conditions = append(conditions, fmt.Sprintf("e.domain = ANY($%d::text[])", len(args)))
	}
	if len(filter.DeviceClasses) > 0 {
		args = append(args, filter.DeviceClasses)
		conditions = append(conditions, fmt.Sprintf("e.device_class = ANY($%d::text[])", len(args)))
	}
	if len(filter.DeviceTypes) > 0 {
		args = append(args, filter.DeviceTypes)
		conditions = append(conditions, fmt.Sprintf("d.device_type = ANY($%d::text[])", len(args)))
	}
	if filter.Controllable != nil {
		args = append(args, *filter.Controllable)
		conditions = append(conditions, fmt.Sprintf("e.controllable = $%d", len(args)))
	}
	if filter.Floor != nil {
		args = append(args, *filter.Floor)
		conditions = append(conditions, fmt.Sprintf("a.floor = $%d", len(args)))
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := `SELECT count(*)
		FROM entities e
		JOIN devices d ON d.id = e.device_id
		LEFT JOIN areas a ON a.id = e.area_id` + where

	total := 0
	if err := r.db.Conn.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return domain.EntityStatePage{}, fmt.Errorf("count entity states: %w", err)
	}

	args = append(args, filter.Limit, filter.Offset)
	pageQuery := `SELECT` + entityStateColumns + `
		FROM entities e
		JOIN devices d ON d.id = e.device_id
		LEFT JOIN areas a ON a.id = e.area_id
		LEFT JOIN twin_state ts ON ts.entity_id = e.id` + where + `
		ORDER BY d.home_id, e.external_entity_id, e.id
		LIMIT $` + strconv.Itoa(len(args)-1) + ` OFFSET $` + strconv.Itoa(len(args))

	rows, err := r.db.Conn.Query(ctx, pageQuery, args...)
	if err != nil {
		return domain.EntityStatePage{}, fmt.Errorf("list entity states: %w", err)
	}
	defer rows.Close()

	entities := []domain.EntityState{}
	for rows.Next() {
		e, err := scanEntityState(rows)
		if err != nil {
			return domain.EntityStatePage{}, fmt.Errorf("list entity states: %w", err)
		}
		entities = append(entities, e)
	}
	if err := rows.Err(); err != nil {
		return domain.EntityStatePage{}, fmt.Errorf("list entity states: %w", err)
	}

	return domain.EntityStatePage{
		Items:  entities,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}, nil
}

func (r *TwinStateRepoImpl) ListEntityStatesByHome(ctx context.Context, homeID string) ([]domain.EntityState, error) {
	return r.listEntityStates(ctx, listEntityStatesByHomeQuery, homeID)
}

func (r *TwinStateRepoImpl) listEntityStates(ctx context.Context, query string, args ...any) ([]domain.EntityState, error) {
	rows, err := r.db.Conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list entity states: %w", err)
	}
	defer rows.Close()

	entities := []domain.EntityState{}
	for rows.Next() {
		e, err := scanEntityState(rows)
		if err != nil {
			return nil, fmt.Errorf("list entity states: %w", err)
		}
		entities = append(entities, e)
	}
	return entities, rows.Err()
}

func (r *TwinStateRepoImpl) GetEntityState(ctx context.Context, entityID string) (domain.EntityState, error) {
	e, err := scanEntityState(r.db.Conn.QueryRow(ctx, getEntityStateQuery, entityID))
	if errors.Is(err, pgx.ErrNoRows) {
		return e, domain.ErrNotFound
	}
	if err != nil {
		return e, fmt.Errorf("get entity state: %w", err)
	}
	return e, nil
}

// Satisfied by both pgx.Row and pgx.Rows
type scanner interface {
	Scan(dest ...any) error
}

func scanEntityState(row scanner) (domain.EntityState, error) {
	var (
		state         domain.EntityState
		valueNum      *float64
		valueText     *string
		attrs         map[string]any
		updatedAt     *time.Time
		prevValueNum  *float64
		prevValueText *string
		prevAttrs     map[string]any
		prevUpdatedAt *time.Time
	)

	err := row.Scan(
		&state.EntityID,
		&state.ExternalEntityID,
		&state.Name,
		&state.Domain,
		&state.Controllable,
		&state.DeviceClass,
		&state.Unit,
		&state.AreaID,
		&state.DeviceID,
		&state.HomeID,
		&state.GatewayID,
		&valueNum,
		&valueText,
		&attrs,
		&updatedAt,
		&prevValueNum,
		&prevValueText,
		&prevAttrs,
		&prevUpdatedAt,
	)
	if err != nil {
		return state, err
	}

	// No twin_state row yet, entity known but unread
	if updatedAt != nil {
		state.State = &domain.StateValue{
			ValueNum:   valueNum,
			ValueText:  valueText,
			Attributes: attrs,
			UpdatedAt:  *updatedAt,
		}
	}

	// Only one accepted reading, nothing before it
	if prevUpdatedAt != nil {
		state.Previous = &domain.StateValue{
			ValueNum:   prevValueNum,
			ValueText:  prevValueText,
			Attributes: prevAttrs,
			UpdatedAt:  *prevUpdatedAt,
		}
	}

	return state, nil
}

// Keeps nil or empty attributes a NULL jsonb instead of {}
//
// # Inputs:
//
//   - reading [dto.Reading] reading with attributes
//
// # Returns:
//
//   - Attributes map or nil for jsonb NULL
func attributesParam(reading dto.Reading) any {
	if len(reading.Attributes) == 0 {
		return nil
	}
	return reading.Attributes
}

// Picks entity name from friendly_name attribute
//
// # Inputs:
//
//   - reading [dto.Reading] reading with attributes
//
// # Returns:
//
//   - friendly_name when set, else external_entity_id
func entityName(reading dto.Reading) string {
	if name, ok := reading.Attributes["friendly_name"].(string); ok && strings.TrimSpace(name) != "" {
		return domain.TruncateRunes(name, 255)
	}
	return reading.ExternalEntityID
}
