package repository

import (
	"context"
	"errors"
	"fmt"

	"IDIG4110/twin-core/internal/domain"

	"github.com/jackc/pgx/v5"
)

// Homes are the catalog, no row level security: which user may see which
// home is the auth service's concern
const (
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
)

// Lists all homes
//
// # Inputs:
//
//   - ctx [context.Context] request context
//
// # Returns:
//
//   - Home summaries
//   - Error on failure
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

// Returns one home by id
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - homeID [string] home id
//
// # Returns:
//
//   - Home summary
//   - Error [domain.ErrNotFound] when the home is unknown, otherwise on failure
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