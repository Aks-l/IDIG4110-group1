package repository

import (
	"context"
	"fmt"

	"IDIG4110/twin-core/internal/domain"
)

// Node deletes, edges are removed explicitly in one transaction
const (
	deleteHomeQuery = `
		DELETE FROM homes WHERE id = $1::uuid
	`

	deleteAreaQuery = `
		DELETE FROM areas WHERE id = $1::uuid
	`

	deleteDeviceQuery = `
		DELETE FROM devices WHERE id = $1::uuid
	`

	deleteEntityQuery = `
		DELETE FROM entities WHERE id = $1::uuid
	`

	deleteAreaRelationsQuery = `
		DELETE FROM twin_relations
		WHERE (from_kind = 'area' AND from_id = $1::uuid)
		   OR (to_kind = 'area' AND to_id = $1::uuid)
	`

	deleteDeviceRelationsQuery = `
		DELETE FROM twin_relations
		WHERE (from_kind = 'device' AND from_id = $1::uuid)
		   OR (to_kind = 'device' AND to_id = $1::uuid)
		   OR (from_kind = 'entity' AND from_id IN (SELECT id FROM entities WHERE device_id = $1::uuid))
		   OR (to_kind = 'entity' AND to_id IN (SELECT id FROM entities WHERE device_id = $1::uuid))
	`

	deleteEntityRelationsQuery = `
		DELETE FROM twin_relations
		WHERE (from_kind = 'entity' AND from_id = $1::uuid)
		   OR (to_kind = 'entity' AND to_id = $1::uuid)
	`
)

// Removes home, everything under it cascades
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - homeID [string] home id
//
// # Returns:
//
//   - ErrNotFound when home missing
func (r *TwinStateRepoImpl) DeleteHome(ctx context.Context, homeID string) error {
	ct, err := r.db.Conn.Exec(ctx, deleteHomeQuery, homeID)
	if err != nil {
		return fmt.Errorf("delete home: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Removes area and its relations, devices and entities keep existing
// with area_id set to null
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - areaID [string] area id
//
// # Returns:
//
//   - ErrNotFound when area missing
func (r *TwinStateRepoImpl) DeleteArea(ctx context.Context, areaID string) error {
	tx, err := r.db.Conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, deleteAreaRelationsQuery, areaID); err != nil {
		return fmt.Errorf("delete area relations: %w", err)
	}

	ct, err := tx.Exec(ctx, deleteAreaQuery, areaID)
	if err != nil {
		return fmt.Errorf("delete area: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return tx.Commit(ctx)
}

// Removes device with its entities, state rows and relations
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - deviceID [string] device id
//
// # Returns:
//
//   - ErrNotFound when device missing
func (r *TwinStateRepoImpl) DeleteDevice(ctx context.Context, deviceID string) error {
	tx, err := r.db.Conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, deleteDeviceRelationsQuery, deviceID); err != nil {
		return fmt.Errorf("delete device relations: %w", err)
	}

	ct, err := tx.Exec(ctx, deleteDeviceQuery, deviceID)
	if err != nil {
		return fmt.Errorf("delete device: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return tx.Commit(ctx)
}

// Removes entity with its state row and relations
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - entityID [string] entity id
//
// # Returns:
//
//   - ErrNotFound when entity missing
func (r *TwinStateRepoImpl) DeleteEntity(ctx context.Context, entityID string) error {
	tx, err := r.db.Conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, deleteEntityRelationsQuery, entityID); err != nil {
		return fmt.Errorf("delete entity relations: %w", err)
	}

	ct, err := tx.Exec(ctx, deleteEntityQuery, entityID)
	if err != nil {
		return fmt.Errorf("delete entity: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return tx.Commit(ctx)
}
