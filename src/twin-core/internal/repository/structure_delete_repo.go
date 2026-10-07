package repository

import (
	"context"
	"fmt"

	"IDIG4110/twin-core/internal/domain"
)

// Node deletes. Relation endpoints are plain uuids (no foreign keys), so
// each delete removes the edges that referenced the node before deleting it,
// in one transaction.
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

// DeleteHome removes a home with everything under it: areas, devices,
// entities, twin_state rows, and relations all cascade on their home or
// entity foreign keys.
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

// DeleteArea removes an area and its relations. Devices and entities keep
// existing: the foreign key sets their area_id to null, so deleting a room
// unassigns instead of destroying them.
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

// DeleteDevice removes a device with its entities and their twin_state rows
// (foreign key cascades), plus every relation that referenced the device or
// any of its entities.
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

// DeleteEntity removes an entity with its twin_state row (foreign key
// cascade) and the relations that referenced it.
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
