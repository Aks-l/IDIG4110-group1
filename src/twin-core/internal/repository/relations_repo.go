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
	query := ""
	switch kind {
	case "area":
		query = `SELECT home_id::text FROM areas WHERE id = $1::uuid`
	case "device":
		query = `SELECT home_id::text FROM devices WHERE id = $1::uuid`
	case "entity":
		query = `
			SELECT d.home_id::text
			FROM entities e
			JOIN devices d ON d.id = e.device_id
			WHERE e.id = $1::uuid
		`
	default:
		return "", fmt.Errorf("%w: unknown node kind %q", domain.ErrBadRequest, kind)
	}

	var homeID string
	err := r.db.Conn.QueryRow(ctx, query, nodeID).Scan(&homeID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("%w: %s %q does not exist", domain.ErrNotFound, kind, nodeID)
	}
	if err != nil {
		return "", fmt.Errorf("resolve %s %s: %w", kind, nodeID, err)
	}
	return homeID, nil
}

func (r *TwinStateRepoImpl) CreateRelation(ctx context.Context, homeID string, req domain.CreateRelationRequest) (domain.Relation, error) {
	var rel domain.Relation
	err := r.db.Conn.QueryRow(ctx, createRelationQuery,
		homeID, req.FromKind, req.FromID, req.ToKind, req.ToID, req.RelationType,
		req.Bidirectional, req.Label, jsonbParam(req.Properties), req.ValidFrom, req.ValidTo,
	).Scan(&rel.ID, &rel.HomeID, &rel.FromKind, &rel.FromID, &rel.ToKind, &rel.ToID,
		&rel.RelationType, &rel.Bidirectional, &rel.Label, &rel.Properties,
		&rel.ValidFrom, &rel.ValidTo)
	if err != nil {
		return domain.Relation{}, fmt.Errorf("create relation: %w", mapWriteError(err))
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
	var rel domain.Relation
	err := r.db.Conn.QueryRow(ctx, getRelationByIDQuery, relationID).
		Scan(&rel.ID, &rel.HomeID, &rel.FromKind, &rel.FromID, &rel.ToKind, &rel.ToID,
			&rel.RelationType, &rel.Bidirectional, &rel.Label, &rel.Properties,
			&rel.ValidFrom, &rel.ValidTo)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Relation{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Relation{}, fmt.Errorf("get relation: %w", err)
	}
	return rel, nil
}

func (r *TwinStateRepoImpl) UpdateRelation(ctx context.Context, relationID string, req domain.UpdateRelationRequest) (domain.Relation, error) {
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

	ct, err := r.db.Conn.Exec(ctx, query, args...)
	if err != nil {
		return domain.Relation{}, fmt.Errorf("update relation: %w", mapWriteError(err))
	}
	if ct.RowsAffected() == 0 {
		return domain.Relation{}, domain.ErrNotFound
	}

	return r.GetRelation(ctx, relationID)
}

func (r *TwinStateRepoImpl) DeleteRelation(ctx context.Context, relationID string) error {
	ct, err := r.db.Conn.Exec(ctx, `DELETE FROM twin_relations WHERE id = $1::uuid`, relationID)
	if err != nil {
		return fmt.Errorf("delete relation: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
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
	var one int
	err := r.db.Conn.QueryRow(ctx, `SELECT 1 FROM homes WHERE id = $1::uuid`, homeID).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("check home: %w", err)
	}

	rows, err := r.db.Conn.Query(ctx, `
		SELECT `+relationColumns+`
		FROM twin_relations
		WHERE home_id = $1::uuid
		ORDER BY created_at, id
	`, homeID)
	if err != nil {
		return nil, fmt.Errorf("list relations: %w", err)
	}
	defer rows.Close()

	relations := []domain.Relation{}
	for rows.Next() {
		var rel domain.Relation
		if err := rows.Scan(&rel.ID, &rel.HomeID, &rel.FromKind, &rel.FromID,
			&rel.ToKind, &rel.ToID, &rel.RelationType, &rel.Bidirectional,
			&rel.Label, &rel.Properties, &rel.ValidFrom, &rel.ValidTo); err != nil {
			return nil, fmt.Errorf("scan relation: %w", err)
		}
		relations = append(relations, rel)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list relations: %w", err)
	}
	return relations, nil
}
