package repository

import (
	"context"
	"fmt"

	"IDIG4110/twin-core/internal/domain"

	"github.com/jackc/pgx/v5"
)

// Row level security scopes areas to app.home_id, WithHome sets it
const listAreasByHomeQuery = `
	SELECT id::text, home_id::text, name, floor, area_type, geometry
	FROM areas
	ORDER BY name
`

// Lists the areas of one home
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - homeID [string] home id
//
// # Returns:
//
//   - Areas of the home
//   - Error on failure
func (r *TwinStateRepoImpl) ListAreasByHome(ctx context.Context, homeID string) ([]domain.Area, error) {
	areas := []domain.Area{}
	err := r.db.WithHome(ctx, homeID, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, listAreasByHomeQuery)
		if err != nil {
			return fmt.Errorf("list areas: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var a domain.Area
			if err := rows.Scan(&a.ID, &a.HomeID, &a.Name, &a.Floor, &a.AreaType, &a.Geometry); err != nil {
				return fmt.Errorf("list areas: %w", err)
			}
			areas = append(areas, a)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return areas, nil
}
