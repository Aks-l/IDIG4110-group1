package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"IDIG4110/twin-core/internal/domain"

	"github.com/jackc/pgx/v5"
)

// Relation queries, endpoints are plain uuids validated via ResolveNodeHome
const (
	relationColumns = `
		id::text,
		home_id::text,
		from_kind,
		from_id::text,
		to_kind,
		to_id::text,
		relation_type,
		bidirectional,
		label,
		properties,
		valid_from,
		valid_to
	`

	createRelationQuery = `
		INSERT INTO twin_relations (home_id, from_kind, from_id, to_kind, to_id,
			relation_type, bidirectional, label, properties, valid_from, valid_to)
		VALUES ($1::uuid, $2, $3::uuid, $4, $5::uuid, $6, $7, $8, $9, $10, $11)
		RETURNING ` + relationColumns + `
	`

	getRelationByIDQuery = `
		SELECT ` + relationColumns + `
		FROM twin_relations
		WHERE id = $1::uuid
	`

	deleteRelationQuery = `
		DELETE FROM twin_relations WHERE id = $1::uuid
	`

	// Row level security scopes the home's rows, no home_id filter needed
	listRelationsByHomeQuery = `
		SELECT ` + relationColumns + `
		FROM twin_relations
		ORDER BY created_at, id
	`
)

// Reports which home a relation endpoint belongs to
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - kind [string] node kind, area | device | entity
//   - nodeID [string] node id
//
// # Returns:
//
//   - Home id of the node
//   - ErrNotFound when node missing, ErrBadRequest on unknown kind
func (r *TwinStateRepoImpl) ResolveNodeHome(ctx context.Context, kind, nodeID string) (string, error) {
	switch kind {
	case nodeKindArea, nodeKindDevice, nodeKindEntity:
	default:
		return "", fmt.Errorf("%w: unknown node kind %q", domain.ErrBadRequest, kind)
	}

	// The node tables are row level secured, the registry resolves without
	// a home context
	homeID, err := r.nodeHome(ctx, nodeID, kind)
	if errors.Is(err, domain.ErrNotFound) {
		return "", fmt.Errorf("%w: %s %q does not exist", domain.ErrNotFound, kind, nodeID)
	}
	if err != nil {
		return "", err
	}
	return homeID, nil
}

func (r *TwinStateRepoImpl) CreateRelation(ctx context.Context, homeID string, req domain.CreateRelationRequest) (domain.Relation, error) {
	var rel domain.Relation
	err := r.db.WithHome(ctx, homeID, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, createRelationQuery,
			homeID, req.FromKind, req.FromID, req.ToKind, req.ToID, req.RelationType,
			req.Bidirectional, req.Label, jsonbParam(req.Properties), req.ValidFrom, req.ValidTo,
		).Scan(&rel.ID, &rel.HomeID, &rel.FromKind, &rel.FromID, &rel.ToKind, &rel.ToID,
			&rel.RelationType, &rel.Bidirectional, &rel.Label, &rel.Properties,
			&rel.ValidFrom, &rel.ValidTo)
		if err != nil {
			return fmt.Errorf("create relation: %w", mapWriteError(err))
		}
		if _, err := tx.Exec(ctx, registerNodeQuery, rel.ID, nodeKindRelation, homeID); err != nil {
			return fmt.Errorf("register relation: %w", err)
		}
		return nil
	})
	if err != nil {
		return domain.Relation{}, err
	}
	return rel, nil
}

// Reads one relation by id, backs the PATCH no-op path
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - relationID [string] relation id
//
// # Returns:
//
//   - Relation
//   - ErrNotFound when missing
func (r *TwinStateRepoImpl) GetRelation(ctx context.Context, relationID string) (domain.Relation, error) {
	homeID, err := r.nodeHome(ctx, relationID, nodeKindRelation)
	if err != nil {
		return domain.Relation{}, err
	}

	var rel domain.Relation
	err = r.db.WithHome(ctx, homeID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, getRelationByIDQuery, relationID).
			Scan(&rel.ID, &rel.HomeID, &rel.FromKind, &rel.FromID, &rel.ToKind, &rel.ToID,
				&rel.RelationType, &rel.Bidirectional, &rel.Label, &rel.Properties,
				&rel.ValidFrom, &rel.ValidTo)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Relation{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Relation{}, fmt.Errorf("get relation: %w", err)
	}
	return rel, nil
}

func (r *TwinStateRepoImpl) UpdateRelation(ctx context.Context, relationID string, req domain.UpdateRelationRequest) (domain.Relation, error) {
	homeID, err := r.nodeHome(ctx, relationID, nodeKindRelation)
	if err != nil {
		return domain.Relation{}, err
	}

	sets := []string{}
	args := []any{relationID}

	if req.RelationType.Set {
		args = append(args, req.RelationType.Value)
		sets = append(sets, fmt.Sprintf("relation_type = $%d", len(args)))
	}
	if req.Bidirectional.Set {
		args = append(args, req.Bidirectional.Value)
		sets = append(sets, fmt.Sprintf("bidirectional = $%d", len(args)))
	}
	if req.Label.Set {
		args = append(args, req.Label.Value)
		sets = append(sets, fmt.Sprintf("label = $%d", len(args)))
	}
	if req.Properties.Set {
		var props map[string]any
		if req.Properties.Value != nil {
			props = *req.Properties.Value
		}
		args = append(args, jsonbParam(props))
		sets = append(sets, fmt.Sprintf("properties = $%d", len(args)))
	}
	if req.ValidFrom.Set {
		args = append(args, req.ValidFrom.Value)
		sets = append(sets, fmt.Sprintf("valid_from = $%d", len(args)))
	}
	if req.ValidTo.Set {
		args = append(args, req.ValidTo.Value)
		sets = append(sets, fmt.Sprintf("valid_to = $%d", len(args)))
	}

	if len(sets) == 0 {
		// No fields to change: return the current row.
		return r.GetRelation(ctx, relationID)
	}

	sets = append(sets, "updated_at = now()")
	query := "UPDATE twin_relations SET " + strings.Join(sets, ", ") +
		" WHERE id = $1::uuid"

	err = r.db.WithHome(ctx, homeID, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("update relation: %w", mapWriteError(err))
		}
		if ct.RowsAffected() == 0 {
			return domain.ErrNotFound
		}
		return nil
	})
	if err != nil {
		return domain.Relation{}, err
	}

	return r.GetRelation(ctx, relationID)
}

func (r *TwinStateRepoImpl) DeleteRelation(ctx context.Context, relationID string) error {
	homeID, err := r.nodeHome(ctx, relationID, nodeKindRelation)
	if err != nil {
		return err
	}

	return r.db.WithHome(ctx, homeID, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, deleteRelationQuery, relationID)
		if err != nil {
			return fmt.Errorf("delete relation: %w", err)
		}
		if ct.RowsAffected() == 0 {
			return domain.ErrNotFound
		}

		if _, err := tx.Exec(ctx, unregisterNodeQuery, relationID, nodeKindRelation); err != nil {
			return fmt.Errorf("unregister relation: %w", err)
		}
		return nil
	})
}

// Returns a home's edges for the graph view, unknown home is ErrNotFound
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - homeID [string] home id
//
// # Returns:
//
//   - Home's relations
//   - ErrNotFound when home missing
func (r *TwinStateRepoImpl) ListRelationsByHome(ctx context.Context, homeID string) ([]domain.Relation, error) {
	if err := r.checkHome(ctx, homeID); err != nil {
		return nil, err
	}

	relations := []domain.Relation{}
	err := r.db.WithHome(ctx, homeID, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, listRelationsByHomeQuery)
		if err != nil {
			return fmt.Errorf("list relations: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var rel domain.Relation
			if err := rows.Scan(&rel.ID, &rel.HomeID, &rel.FromKind, &rel.FromID,
				&rel.ToKind, &rel.ToID, &rel.RelationType, &rel.Bidirectional,
				&rel.Label, &rel.Properties, &rel.ValidFrom, &rel.ValidTo); err != nil {
				return fmt.Errorf("scan relation: %w", err)
			}
			relations = append(relations, rel)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return relations, nil
}
