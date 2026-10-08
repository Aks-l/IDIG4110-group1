package repository

import (
	"context"
	"errors"
	"fmt"

	"IDIG4110/twin-core/internal/domain"

	"github.com/jackc/pgx/v5"
)

// Registry kinds, must match the node_registry kind CHECK constraint
const (
	nodeKindArea     = "area"
	nodeKindDevice   = "device"
	nodeKindEntity   = "entity"
	nodeKindRelation = "relation"
)

// Readings and bare-id API calls resolve their home before the home tables,
// whose row level security hides everything outside app.home_id. The
// registries hold routing data only, which user may access which home is
// the auth service's concern.
const (
	routeDeviceQuery = `
		SELECT home_id::text, device_id::text
		FROM device_registry
		WHERE gateway_id = $1::uuid AND external_id = $2
	`

	registerDeviceQuery = `
		INSERT INTO device_registry (gateway_id, external_id, home_id, device_id)
		VALUES ($1::uuid, $2, $3::uuid, $4::uuid)
		ON CONFLICT (gateway_id, external_id) DO NOTHING
	`

	// Entities are addressed by bare id in the API
	ensureEntityRegistryQuery = `
		INSERT INTO node_registry (id, kind, home_id)
		VALUES ($1::uuid, 'entity', $2::uuid)
		ON CONFLICT (id, kind) DO NOTHING
	`

	resolveNodeHomeQuery = `
		SELECT home_id::text
		FROM node_registry
		WHERE id = $1::uuid AND kind = $2
	`

	registerNodeQuery = `
		INSERT INTO node_registry (id, kind, home_id)
		VALUES ($1::uuid, $2, $3::uuid)
		ON CONFLICT (id, kind) DO NOTHING
	`

	unregisterNodeQuery = `
		DELETE FROM node_registry WHERE id = $1::uuid AND kind = $2
	`

	checkHomeQuery = `
		SELECT 1 FROM homes WHERE id = $1::uuid
	`
)

// Resolves a bare node id to its home through the registry
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - id [string] node id
//   - kind [string] registry kind, one of the nodeKind constants
//
// # Returns:
//
//   - Home id
//   - Error [domain.ErrNotFound] when the node is unknown, otherwise on failure
func (r *TwinStateRepoImpl) nodeHome(ctx context.Context, id, kind string) (string, error) {
	var homeID string
	err := r.db.Conn.QueryRow(ctx, resolveNodeHomeQuery, id, kind).Scan(&homeID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("resolve node home: %w", err)
	}
	return homeID, nil
}

// Reports [domain.ErrNotFound] when the home does not exist, homes carry no
// row level security so this works without a home context
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - homeID [string] home id
//
// # Returns:
//
//   - Error [domain.ErrNotFound] when the home is unknown, otherwise on failure
func (r *TwinStateRepoImpl) checkHome(ctx context.Context, homeID string) error {
	var one int
	err := r.db.Conn.QueryRow(ctx, checkHomeQuery, homeID).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("check home: %w", err)
	}
	return nil
}