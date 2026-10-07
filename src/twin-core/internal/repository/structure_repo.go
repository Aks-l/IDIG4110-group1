package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"IDIG4110/shared/pgerrors"
	"IDIG4110/twin-core/internal/domain"

	"github.com/jackc/pgx/v5"
)

const (
	// Structure API write queries, auto-provisioning lives in twin_state_repo.go

	createHomeQuery = `
		INSERT INTO homes (name, address, timezone)
		VALUES ($1, $2, $3)
		RETURNING id::text, name, address, timezone
	`

	createAreaQuery = `
		INSERT INTO areas (home_id, name, floor, area_type, geometry)
		VALUES ($1::uuid, $2, $3, $4, $5)
		RETURNING id::text, home_id::text, name, floor, area_type, geometry
	`

	getAreaByIDQuery = `
		SELECT id::text, home_id::text, name, floor, area_type, geometry
		FROM areas
		WHERE id = $1::uuid
	`

	createDeviceQuery = `
		INSERT INTO devices (home_id, area_id, gateway_id, external_id, name,
			manufacturer, model, sw_version, device_type)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6, $7, $8, $9)
		RETURNING id::text, home_id::text, area_id::text, gateway_id::text,
			external_id, name, manufacturer, model, sw_version, device_type
	`

	getDeviceByIDQuery = `
		SELECT id::text, home_id::text, area_id::text, gateway_id::text,
			external_id, name, manufacturer, model, sw_version, device_type
		FROM devices
		WHERE id = $1::uuid
	`

	createEntityQuery = `
		INSERT INTO entities (device_id, area_id, external_entity_id, name,
			domain, device_class, unit, controllable, command_map, state_ttl_seconds)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id::text
	`

	areaColumns = `
		id::text,
		home_id::text,
		name,
		floor,
		area_type,
		geometry
	`

	deviceColumns = `
		id::text,
		home_id::text,
		area_id::text,
		gateway_id::text,
		external_id,
		name,
		manufacturer,
		model,
		sw_version,
		device_type
	`
)

func (r *TwinStateRepoImpl) CreateHome(ctx context.Context, req domain.CreateHomeRequest) (domain.HomeSummary, error) {
	var h domain.HomeSummary
	err := r.db.Conn.QueryRow(ctx, createHomeQuery, req.Name, req.Address, req.Timezone).
		Scan(&h.ID, &h.Name, &h.Address, &h.Timezone)
	if err != nil {
		return h, fmt.Errorf("create home: %w", mapWriteError(err))
	}
	return h, nil
}

func (r *TwinStateRepoImpl) UpdateHome(ctx context.Context, homeID string, req domain.UpdateHomeRequest) (domain.HomeSummary, error) {
	sets := []string{}
	args := []any{homeID}

	if req.Name.Set {
		args = append(args, req.Name.Value)
		sets = append(sets, fmt.Sprintf("name = $%d", len(args)))
	}
	if req.Address.Set {
		args = append(args, req.Address.Value)
		sets = append(sets, fmt.Sprintf("address = $%d", len(args)))
	}
	if req.Timezone.Set {
		args = append(args, req.Timezone.Value)
		sets = append(sets, fmt.Sprintf("timezone = $%d", len(args)))
	}

	if len(sets) == 0 {
		// No fields to change: return the current row.
		return r.GetHome(ctx, homeID)
	}

	sets = append(sets, "updated_at = now()")
	query := "UPDATE homes SET " + strings.Join(sets, ", ") +
		" WHERE id = $1::uuid RETURNING id::text, name, address, timezone"

	var h domain.HomeSummary
	err := r.db.Conn.QueryRow(ctx, query, args...).
		Scan(&h.ID, &h.Name, &h.Address, &h.Timezone)
	if errors.Is(err, pgx.ErrNoRows) {
		return h, domain.ErrNotFound
	}
	if err != nil {
		return h, fmt.Errorf("update home: %w", mapWriteError(err))
	}
	return h, nil
}

func (r *TwinStateRepoImpl) CreateArea(ctx context.Context, req domain.CreateAreaRequest) (domain.Area, error) {
	var a domain.Area
	err := r.db.Conn.QueryRow(ctx, createAreaQuery,
		req.HomeID, req.Name, req.Floor, req.AreaType, jsonbArg(req.Geometry)).
		Scan(&a.ID, &a.HomeID, &a.Name, &a.Floor, &a.AreaType, &a.Geometry)
	if err != nil {
		return a, fmt.Errorf("create area: %w", mapWriteError(err))
	}
	return a, nil
}

func (r *TwinStateRepoImpl) UpdateArea(ctx context.Context, areaID string, req domain.UpdateAreaRequest) (domain.Area, error) {
	sets := []string{}
	args := []any{areaID}

	if req.Name.Set {
		args = append(args, req.Name.Value)
		sets = append(sets, fmt.Sprintf("name = $%d", len(args)))
	}
	if req.Floor.Set {
		args = append(args, req.Floor.Value)
		sets = append(sets, fmt.Sprintf("floor = $%d", len(args)))
	}
	if req.AreaType.Set {
		args = append(args, req.AreaType.Value)
		sets = append(sets, fmt.Sprintf("area_type = $%d", len(args)))
	}
	if req.Geometry.Set {
		if req.Geometry.Value == nil {
			args = append(args, nil)
		} else {
			args = append(args, jsonbArg(*req.Geometry.Value))
		}
		sets = append(sets, fmt.Sprintf("geometry = $%d", len(args)))
	}

	if len(sets) == 0 {
		// No fields to change: return the current row.
		return r.GetArea(ctx, areaID)
	}

	sets = append(sets, "updated_at = now()")
	query := "UPDATE areas SET " + strings.Join(sets, ", ") +
		" WHERE id = $1::uuid RETURNING " + areaColumns

	var a domain.Area
	err := r.db.Conn.QueryRow(ctx, query, args...).
		Scan(&a.ID, &a.HomeID, &a.Name, &a.Floor, &a.AreaType, &a.Geometry)
	if errors.Is(err, pgx.ErrNoRows) {
		return a, domain.ErrNotFound
	}
	if err != nil {
		return a, fmt.Errorf("update area: %w", mapWriteError(err))
	}
	return a, nil
}

// Fetches one area by id, backs empty patch no-op response
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - areaID [string] area id
//
// # Returns:
//
//   - Area
//   - ErrNotFound when missing
func (r *TwinStateRepoImpl) GetArea(ctx context.Context, areaID string) (domain.Area, error) {
	var a domain.Area
	err := r.db.Conn.QueryRow(ctx, getAreaByIDQuery, areaID).
		Scan(&a.ID, &a.HomeID, &a.Name, &a.Floor, &a.AreaType, &a.Geometry)
	if errors.Is(err, pgx.ErrNoRows) {
		return a, domain.ErrNotFound
	}
	if err != nil {
		return a, fmt.Errorf("get area: %w", err)
	}
	return a, nil
}

func (r *TwinStateRepoImpl) CreateDevice(ctx context.Context, req domain.CreateDeviceRequest) (domain.Device, error) {
	var d domain.Device
	err := r.db.Conn.QueryRow(ctx, createDeviceQuery,
		req.HomeID, req.AreaID, req.GatewayID, req.ExternalID, req.Name,
		req.Manufacturer, req.Model, req.SwVersion, req.DeviceType).
		Scan(&d.ID, &d.HomeID, &d.AreaID, &d.GatewayID, &d.ExternalID, &d.Name,
			&d.Manufacturer, &d.Model, &d.SwVersion, &d.DeviceType)
	if err != nil {
		return d, fmt.Errorf("create device: %w", mapWriteError(err))
	}
	return d, nil
}

func (r *TwinStateRepoImpl) UpdateDevice(ctx context.Context, deviceID string, req domain.UpdateDeviceRequest) (domain.Device, error) {
	sets := []string{}
	args := []any{deviceID}

	if req.Name.Set {
		args = append(args, req.Name.Value)
		sets = append(sets, fmt.Sprintf("name = $%d", len(args)))
	}
	if req.HomeID.Set {
		args = append(args, req.HomeID.Value)
		sets = append(sets, fmt.Sprintf("home_id = $%d::uuid", len(args)))
	}
	if req.AreaID.Set {
		args = append(args, req.AreaID.Value)
		sets = append(sets, fmt.Sprintf("area_id = $%d::uuid", len(args)))
	}
	if req.Manufacturer.Set {
		args = append(args, req.Manufacturer.Value)
		sets = append(sets, fmt.Sprintf("manufacturer = $%d", len(args)))
	}
	if req.Model.Set {
		args = append(args, req.Model.Value)
		sets = append(sets, fmt.Sprintf("model = $%d", len(args)))
	}
	if req.SwVersion.Set {
		args = append(args, req.SwVersion.Value)
		sets = append(sets, fmt.Sprintf("sw_version = $%d", len(args)))
	}
	if req.DeviceType.Set {
		args = append(args, req.DeviceType.Value)
		sets = append(sets, fmt.Sprintf("device_type = $%d", len(args)))
	}

	if len(sets) == 0 {
		// No fields to change: return the current row.
		return r.GetDevice(ctx, deviceID)
	}

	sets = append(sets, "updated_at = now()")
	query := "UPDATE devices SET " + strings.Join(sets, ", ") +
		" WHERE id = $1::uuid RETURNING " + deviceColumns

	var d domain.Device
	err := r.db.Conn.QueryRow(ctx, query, args...).
		Scan(&d.ID, &d.HomeID, &d.AreaID, &d.GatewayID, &d.ExternalID, &d.Name,
			&d.Manufacturer, &d.Model, &d.SwVersion, &d.DeviceType)
	if errors.Is(err, pgx.ErrNoRows) {
		return d, domain.ErrNotFound
	}
	if err != nil {
		return d, fmt.Errorf("update device: %w", mapWriteError(err))
	}
	return d, nil
}

// Fetches one device by id, backs empty patch no-op response
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - deviceID [string] device id
//
// # Returns:
//
//   - Device
//   - ErrNotFound when missing
func (r *TwinStateRepoImpl) GetDevice(ctx context.Context, deviceID string) (domain.Device, error) {
	var d domain.Device
	err := r.db.Conn.QueryRow(ctx, getDeviceByIDQuery, deviceID).
		Scan(&d.ID, &d.HomeID, &d.AreaID, &d.GatewayID, &d.ExternalID, &d.Name,
			&d.Manufacturer, &d.Model, &d.SwVersion, &d.DeviceType)
	if errors.Is(err, pgx.ErrNoRows) {
		return d, domain.ErrNotFound
	}
	if err != nil {
		return d, fmt.Errorf("get device: %w", err)
	}
	return d, nil
}

func (r *TwinStateRepoImpl) CreateEntity(ctx context.Context, req domain.CreateEntityRequest) (domain.EntityState, error) {
	var entityID string
	err := r.db.Conn.QueryRow(ctx, createEntityQuery,
		req.DeviceID, req.AreaID, req.ExternalEntityID, req.Name,
		req.Domain, req.DeviceClass, req.Unit, req.Controllable,
		jsonbParam(req.CommandMap), req.StateTTLSeconds).
		Scan(&entityID)
	if err != nil {
		return domain.EntityState{}, fmt.Errorf("create entity: %w", mapWriteError(err))
	}

	return r.GetEntityState(ctx, entityID)
}

func (r *TwinStateRepoImpl) UpdateEntity(ctx context.Context, entityID string, req domain.UpdateEntityRequest) (domain.EntityState, error) {
	sets := []string{}
	args := []any{entityID}

	if req.Name.Set {
		args = append(args, req.Name.Value)
		sets = append(sets, fmt.Sprintf("name = $%d", len(args)))
	}
	if req.AreaID.Set {
		args = append(args, req.AreaID.Value)
		sets = append(sets, fmt.Sprintf("area_id = $%d::uuid", len(args)))
	}
	if req.DeviceClass.Set {
		args = append(args, req.DeviceClass.Value)
		sets = append(sets, fmt.Sprintf("device_class = $%d", len(args)))
	}
	if req.Unit.Set {
		args = append(args, req.Unit.Value)
		sets = append(sets, fmt.Sprintf("unit = $%d", len(args)))
	}
	if req.Controllable.Set {
		args = append(args, req.Controllable.Value)
		sets = append(sets, fmt.Sprintf("controllable = $%d", len(args)))
	}
	if req.CommandMap.Set {
		args = append(args, req.CommandMap.Value)
		sets = append(sets, fmt.Sprintf("command_map = $%d", len(args)))
	}
	if req.StateTTLSeconds.Set {
		args = append(args, req.StateTTLSeconds.Value)
		sets = append(sets, fmt.Sprintf("state_ttl_seconds = $%d", len(args)))
	}

	if len(sets) == 0 {
		// No fields to change: return the current row.
		return r.GetEntityState(ctx, entityID)
	}

	sets = append(sets, "updated_at = now()")
	query := "UPDATE entities SET " + strings.Join(sets, ", ") +
		" WHERE id = $1::uuid RETURNING id::text"

	var id string
	err := r.db.Conn.QueryRow(ctx, query, args...).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.EntityState{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.EntityState{}, fmt.Errorf("update entity: %w", mapWriteError(err))
	}

	return r.GetEntityState(ctx, id)
}

// Makes decoded JSON insertable into jsonb column
// nil stays NULL, maps pass through, other shapes marshaled
//
// # Inputs:
//
//   - v [any] decoded JSON value
//
// # Returns:
//
//   - Value insertable into jsonb
func jsonbArg(v any) any {
	switch t := v.(type) {
	case nil:
		return nil
	case map[string]any:
		return t
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return v // let pgx surface the encoding error
		}
		return b
	}
}

// Keeps nil or empty maps a NULL jsonb instead of JSON null
//
// # Inputs:
//
//   - m [map[string]any] map to insert
//
// # Returns:
//
//   - Map or nil for jsonb NULL
func jsonbParam(m map[string]any) any {
	if len(m) == 0 {
		return nil
	}
	return m
}

// Maps postgres constraint violations onto domain sentinels
// unique -> conflict, missing reference -> not found, invalid value -> bad request
//
// # Inputs:
//
//   - err [error] error to map
//
// # Returns:
//
//   - Wrapped sentinel or err unchanged
func mapWriteError(err error) error {
	switch {
	case pgerrors.IsUniqueViolation(err):
		return fmt.Errorf("%w: %s", domain.ErrConflict, pgerrors.Message(err))
	case pgerrors.IsForeignKeyViolation(err):
		return fmt.Errorf("%w: %s", domain.ErrNotFound, pgerrors.Message(err))
	case pgerrors.IsNotNullViolation(err),
		pgerrors.IsStringValueTooLong(err),
		pgerrors.IsCheckViolation(err):
		return fmt.Errorf("%w: %s", domain.ErrBadRequest, pgerrors.Message(err))
	}
	return err
}
