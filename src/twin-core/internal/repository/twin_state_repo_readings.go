package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"IDIG4110/shared/dto"
	"IDIG4110/twin-core/internal/domain"

	"github.com/jackc/pgx/v5"
)

// Write path, auto-provisions placeholder device and entity on first sight
const (
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
		INSERT INTO entities (home_id, device_id, external_entity_id, name, domain, device_class, unit)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7)
		ON CONFLICT (device_id, external_entity_id) DO NOTHING
		RETURNING id::text
	`

	getEntityByDeviceQuery = `
		SELECT id::text FROM entities
		WHERE device_id = $1::uuid AND external_entity_id = $2
	`

	// Last-write-wins on reading time, superseded values move to previous_*
	upsertStateQuery = `
		INSERT INTO twin_state (entity_id, home_id, value_num, value_text, attributes, updated_at)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6)
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
)

// Applies one reading to twin_state
// routes through device_registry, auto-provisions placeholder device and
// entity in the default home on first sight, every statement runs under
// the row level security home context
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - reading [dto.Reading] the reading to apply
//
// # Returns:
//
//   - Error on route, provision or upsert failure
func (r *TwinStateRepoImpl) ApplyReading(ctx context.Context, reading dto.Reading) error {
	tx, err := r.db.Conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	routedHome, routedDevice := "", ""
	provisioned := false
	err = tx.QueryRow(ctx, routeDeviceQuery, reading.GatewayID, reading.ExternalEntityID).Scan(&routedHome, &routedDevice)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		// First reading for this (gateway, external id), the default home
		// takes the placeholder device and entity
		if _, err := tx.Exec(ctx, ensureDefaultHomeQuery, domain.DefaultHomeID); err != nil {
			return fmt.Errorf("ensure default home: %w", err)
		}
		routedHome = domain.DefaultHomeID
		provisioned = true
	case err != nil:
		return fmt.Errorf("route reading: %w", err)
	}

	// Scope the rest of the transaction to the routed home, same mechanism
	// as db.WithHome for statements that already own the transaction
	if _, err := tx.Exec(ctx, "SELECT set_config('app.home_id', $1, true)", routedHome); err != nil {
		return fmt.Errorf("set home context: %w", err)
	}

	if routedDevice == "" {
		if routedDevice, err = r.ensureDevice(ctx, tx, reading); err != nil {
			return err
		}
	}

	entityID, err := r.ensureEntity(ctx, tx, routedDevice, routedHome, reading)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, upsertStateQuery, entityID, routedHome, reading.ValueNum, reading.ValueText, attributesParam(reading), reading.Timestamp); err != nil {
		return fmt.Errorf("upsert twin state: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	if provisioned {
		slog.Info("auto-provisioned placeholder entity",
			"gateway_id", reading.GatewayID,
			"external_entity_id", reading.ExternalEntityID,
			"entity_id", entityID,
		)
	}
	return nil
}

// Registers the placeholder device for a first reading, returns its id
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - tx [pgx.Tx] transaction under the default home context
//   - reading [dto.Reading] the reading being applied
//
// # Returns:
//
//   - Device id
//   - Error on failure
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

	if _, err := tx.Exec(ctx, registerDeviceQuery, reading.GatewayID, reading.ExternalEntityID, domain.DefaultHomeID, deviceID); err != nil {
		return "", fmt.Errorf("register device: %w", err)
	}
	// Bare-id API calls (PATCH/DELETE /devices/{id}) resolve their home
	// through node_registry; without this row every auto-provisioned
	// device 404s there, unlike explicitly created devices.
	if _, err := tx.Exec(ctx, registerNodeQuery, deviceID, nodeKindDevice, domain.DefaultHomeID); err != nil {
		return "", fmt.Errorf("register device node: %w", err)
	}
	return deviceID, nil
}

// Registers the entity for a reading when missing, returns its id
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - tx [pgx.Tx] transaction under the reading's home context
//   - deviceID [string] device id
//   - homeID [string] home id
//   - reading [dto.Reading] the reading being applied
//
// # Returns:
//
//   - Entity id
//   - Error on failure
func (r *TwinStateRepoImpl) ensureEntity(ctx context.Context, tx pgx.Tx, deviceID, homeID string, reading dto.Reading) (string, error) {
	deviceClass, unit := reading.DeviceClass, reading.Unit

	var entityID string
	err := tx.QueryRow(
		ctx,
		insertEntityQuery,
		homeID,
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

	if _, err := tx.Exec(ctx, ensureEntityRegistryQuery, entityID, homeID); err != nil {
		return "", fmt.Errorf("register entity: %w", err)
	}
	return entityID, nil
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
