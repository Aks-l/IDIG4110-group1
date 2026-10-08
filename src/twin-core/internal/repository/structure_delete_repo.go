package repository

import (
	"context"
	"fmt"

	"IDIG4110/twin-core/internal/domain"

	"github.com/jackc/pgx/v5"
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

	// Registry rows outlive nothing, the entities subquery must run before
	// the device delete cascades them away
	deleteDeviceRegistryQuery = `
		DELETE FROM node_registry
		WHERE (id = $1::uuid AND kind = 'device')
		   OR (kind = 'entity' AND id IN (SELECT id FROM entities WHERE device_id = $1::uuid))
	`

	unregisterDeviceRouteQuery = `
		DELETE FROM device_registry WHERE device_id = $1::uuid
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
	homeID, err := r.nodeHome(ctx, areaID, nodeKindArea)
	if err != nil {
		return err
	}

	return r.db.WithHome(ctx, homeID, func(tx pgx.Tx) error {
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

		if _, err := tx.Exec(ctx, unregisterNodeQuery, areaID, nodeKindArea); err != nil {
			return fmt.Errorf("unregister area: %w", err)
		}
		return nil
	})
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
	homeID, err := r.nodeHome(ctx, deviceID, nodeKindDevice)
	if err != nil {
		return err
	}

	return r.db.WithHome(ctx, homeID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, deleteDeviceRelationsQuery, deviceID); err != nil {
			return fmt.Errorf("delete device relations: %w", err)
		}

		if _, err := tx.Exec(ctx, deleteDeviceRegistryQuery, deviceID); err != nil {
			return fmt.Errorf("delete device registry: %w", err)
		}
		if _, err := tx.Exec(ctx, unregisterDeviceRouteQuery, deviceID); err != nil {
			return fmt.Errorf("unregister device route: %w", err)
		}

		ct, err := tx.Exec(ctx, deleteDeviceQuery, deviceID)
		if err != nil {
			return fmt.Errorf("delete device: %w", err)
		}
		if ct.RowsAffected() == 0 {
			return domain.ErrNotFound
		}
		return nil
	})
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
	homeID, err := r.nodeHome(ctx, entityID, nodeKindEntity)
	if err != nil {
		return err
	}

	return r.db.WithHome(ctx, homeID, func(tx pgx.Tx) error {
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

		if _, err := tx.Exec(ctx, unregisterNodeQuery, entityID, nodeKindEntity); err != nil {
			return fmt.Errorf("unregister entity: %w", err)
		}
		return nil
	})
}
