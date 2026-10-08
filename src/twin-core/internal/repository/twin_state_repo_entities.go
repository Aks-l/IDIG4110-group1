package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"IDIG4110/twin-core/internal/domain"

	"github.com/jackc/pgx/v5"
)

// uuid columns cast to text, scans into Go strings
const entityStateColumns = `
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

const getEntityStateQuery = `
	SELECT` + entityStateColumns + `
	FROM entities e
	JOIN devices d ON d.id = e.device_id
	LEFT JOIN twin_state ts ON ts.entity_id = e.id
	WHERE e.id = $1::uuid
`

// Returns one page of the entity list
// filters combine with AND, zero filter lists everything; a home filter
// scopes count and page to that home's row level security context, without
// one the rows are collected per home and merged, they are visible only
// inside their own home transaction
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

	if filter.HomeID != "" {
		// One home, count and page run under its row level security context
		page := domain.EntityStatePage{Items: []domain.EntityState{}, Limit: filter.Limit, Offset: filter.Offset}
		err := r.db.WithHome(ctx, filter.HomeID, func(tx pgx.Tx) error {
			countQuery := `SELECT count(*)
				FROM entities e
				JOIN devices d ON d.id = e.device_id
				LEFT JOIN areas a ON a.id = e.area_id` + where

			total := 0
			if err := tx.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
				return fmt.Errorf("count entity states: %w", err)
			}
			page.Total = total

			pageArgs := append(append([]any{}, args...), filter.Limit, filter.Offset)
			pageQuery := `SELECT` + entityStateColumns + `
				FROM entities e
				JOIN devices d ON d.id = e.device_id
				LEFT JOIN areas a ON a.id = e.area_id
				LEFT JOIN twin_state ts ON ts.entity_id = e.id` + where + `
				ORDER BY e.external_entity_id
				LIMIT $` + strconv.Itoa(len(pageArgs)-1) + ` OFFSET $` + strconv.Itoa(len(pageArgs))

			rows, err := tx.Query(ctx, pageQuery, pageArgs...)
			if err != nil {
				return fmt.Errorf("list entity states: %w", err)
			}
			defer rows.Close()

			for rows.Next() {
				e, err := scanEntityState(rows)
				if err != nil {
					return fmt.Errorf("list entity states: %w", err)
				}
				page.Items = append(page.Items, e)
			}
			return rows.Err()
		})
		if err != nil {
			return domain.EntityStatePage{}, err
		}
		return page, nil
	}

	// Cross-home: collect per home and merge the page
	homes, err := r.ListHomes(ctx)
	if err != nil {
		return domain.EntityStatePage{}, err
	}

	items := []domain.EntityState{}
	total := 0
	for _, h := range homes {
		homeItems, err := r.listEntityStatesInHome(ctx, h.ID, where, args)
		if err != nil {
			return domain.EntityStatePage{}, err
		}
		items = append(items, homeItems...)
		total += len(homeItems)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].HomeID != items[j].HomeID {
			return items[i].HomeID < items[j].HomeID
		}
		if items[i].ExternalEntityID != items[j].ExternalEntityID {
			return items[i].ExternalEntityID < items[j].ExternalEntityID
		}
		return items[i].EntityID < items[j].EntityID
	})

	start := filter.Offset
	if start > len(items) {
		start = len(items)
	}
	end := start + filter.Limit
	if end > len(items) {
		end = len(items)
	}

	return domain.EntityStatePage{
		Items:  items[start:end],
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}, nil
}

// Lists the entity states of one home
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - homeID [string] home id
//
// # Returns:
//
//   - Entity states of the home
//   - Error on failure
func (r *TwinStateRepoImpl) ListEntityStatesByHome(ctx context.Context, homeID string) ([]domain.EntityState, error) {
	return r.listEntityStatesInHome(ctx, homeID, "", nil)
}

// Runs the filtered entity state query inside one home transaction, row
// level security hides every other home
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - homeID [string] home id
//   - where [string] SQL WHERE clause from the entity filter
//   - args [][any] filter arguments
//
// # Returns:
//
//   - Entity states of the home
//   - Error on failure
func (r *TwinStateRepoImpl) listEntityStatesInHome(ctx context.Context, homeID, where string, args []any) ([]domain.EntityState, error) {
	entities := []domain.EntityState{}
	err := r.db.WithHome(ctx, homeID, func(tx pgx.Tx) error {
		query := `SELECT` + entityStateColumns + `
			FROM entities e
			JOIN devices d ON d.id = e.device_id
			LEFT JOIN areas a ON a.id = e.area_id
			LEFT JOIN twin_state ts ON ts.entity_id = e.id` + where + `
			ORDER BY e.external_entity_id`

		rows, err := tx.Query(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("list entity states: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			e, err := scanEntityState(rows)
			if err != nil {
				return fmt.Errorf("list entity states: %w", err)
			}
			entities = append(entities, e)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return entities, nil
}

// Returns the state of one entity, its home comes from the registry
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - entityID [string] entity id
//
// # Returns:
//
//   - Entity state
//   - Error [domain.ErrNotFound] when the entity is unknown, otherwise on failure
func (r *TwinStateRepoImpl) GetEntityState(ctx context.Context, entityID string) (domain.EntityState, error) {
	homeID, err := r.nodeHome(ctx, entityID, nodeKindEntity)
	if err != nil {
		return domain.EntityState{}, err
	}

	var e domain.EntityState
	err = r.db.WithHome(ctx, homeID, func(tx pgx.Tx) error {
		var err error
		e, err = scanEntityState(tx.QueryRow(ctx, getEntityStateQuery, entityID))
		return err
	})
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
