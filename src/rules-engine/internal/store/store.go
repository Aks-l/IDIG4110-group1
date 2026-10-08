// Package store reads and writes rules, incidents and notifications in
// rules_db.
package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"IDIG4110/rules-engine/internal/rule"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type Incident struct {
	ID             string         `json:"id"`
	RuleID         string         `json:"rule_id"`
	HomeID         string         `json:"home_id"`
	Severity       string         `json:"severity"`
	Status         string         `json:"status"`
	Message        string         `json:"message"`
	Context        map[string]any `json:"context"`
	TriggeredAt    time.Time      `json:"triggered_at"`
	AcknowledgedAt *time.Time     `json:"acknowledged_at"`
	ResolvedAt     *time.Time     `json:"resolved_at"`
}

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Ping checks that rules_db answers; used by the API readiness probe.
//
// # Inputs:
//
// - ctx: [context.Context] request context
//
// # Returns:
//
// - error: non-nil if the ping failed
func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// --- rules ----------------------------------------------------------------

const ruleColumns = `id, home_id, name, coalesce(description, ''), enabled, priority,
	trigger_config, action_config, fire_count, last_fired_at, created_at, updated_at`

func scanRule(row pgx.Row) (rule.Rule, error) {
	var r rule.Rule
	var triggerJSON, actionJSON []byte
	err := row.Scan(&r.ID, &r.HomeID, &r.Name, &r.Description, &r.Enabled, &r.Priority,
		&triggerJSON, &actionJSON, &r.FireCount, &r.LastFiredAt, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return r, ErrNotFound
	}
	if err != nil {
		return r, err
	}
	return r, r.ApplyConfigs(triggerJSON, actionJSON)
}

// ListRules returns the rules of one home, or of every home when homeID is
// empty, highest priority first.
func (s *Store) ListRules(ctx context.Context, homeID string) ([]rule.Rule, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT `+ruleColumns+` FROM rules
		 WHERE $1 = '' OR home_id::text = $1
		 ORDER BY priority DESC, name`, homeID)
	if err != nil {
		return nil, fmt.Errorf("listing rules: %w", err)
	}
	defer rows.Close()

	rules := []rule.Rule{}
	for rows.Next() {
		r, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

func (s *Store) GetRule(ctx context.Context, id string) (rule.Rule, error) {
	return scanRule(s.pool.QueryRow(ctx, `SELECT `+ruleColumns+` FROM rules WHERE id = $1`, id))
}

func (s *Store) CreateRule(ctx context.Context, r rule.Rule) (rule.Rule, error) {
	triggerJSON, actionJSON, err := configs(r)
	if err != nil {
		return r, err
	}
	return scanRule(s.pool.QueryRow(ctx,
		`INSERT INTO rules (home_id, name, description, enabled, priority, trigger_config, action_config)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING `+ruleColumns,
		r.HomeID, r.Name, r.Description, r.Enabled, r.Priority, triggerJSON, actionJSON))
}

func (s *Store) UpdateRule(ctx context.Context, r rule.Rule) (rule.Rule, error) {
	triggerJSON, actionJSON, err := configs(r)
	if err != nil {
		return r, err
	}
	return scanRule(s.pool.QueryRow(ctx,
		`UPDATE rules SET home_id = $2, name = $3, description = $4, enabled = $5, priority = $6,
		     trigger_config = $7, action_config = $8, updated_at = now()
		 WHERE id = $1
		 RETURNING `+ruleColumns,
		r.ID, r.HomeID, r.Name, r.Description, r.Enabled, r.Priority, triggerJSON, actionJSON))
}

func (s *Store) SetRuleEnabled(ctx context.Context, id string, enabled bool) (rule.Rule, error) {
	return scanRule(s.pool.QueryRow(ctx,
		`UPDATE rules SET enabled = $2, updated_at = now() WHERE id = $1 RETURNING `+ruleColumns,
		id, enabled))
}

// DeleteRule removes a rule. A rule that has raised incidents cannot be
// deleted (incidents.rule_id), so the incident history stays intact; disable
// it instead.
func (s *Store) DeleteRule(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM rules WHERE id = $1`, id)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return fmt.Errorf("%w: rule has incidents; disable it instead", ErrConflict)
	}
	if err != nil {
		return fmt.Errorf("deleting rule: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) RecordFiring(ctx context.Context, ruleID string, at time.Time) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE rules SET fire_count = fire_count + 1, last_fired_at = $2 WHERE id = $1`, ruleID, at)
	return err
}

func configs(r rule.Rule) ([]byte, []byte, error) {
	triggerJSON, err := r.TriggerConfig()
	if err != nil {
		return nil, nil, err
	}
	actionJSON, err := r.ActionConfig()
	return triggerJSON, actionJSON, err
}

// --- incidents ------------------------------------------------------------

const incidentColumns = `id, rule_id, home_id, severity, status, message, context,
	triggered_at, acknowledged_at, resolved_at`

func scanIncident(row pgx.Row) (Incident, error) {
	var inc Incident
	var contextJSON []byte
	err := row.Scan(&inc.ID, &inc.RuleID, &inc.HomeID, &inc.Severity, &inc.Status, &inc.Message,
		&contextJSON, &inc.TriggeredAt, &inc.AcknowledgedAt, &inc.ResolvedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return inc, ErrNotFound
	}
	if err != nil {
		return inc, err
	}
	if len(contextJSON) > 0 {
		err = json.Unmarshal(contextJSON, &inc.Context)
	}
	return inc, err
}

// CreateIncident stores an incident and one pending notification per
// channel, in one transaction.
func (s *Store) CreateIncident(ctx context.Context, inc Incident, channels []string) (Incident, error) {
	contextJSON, err := json.Marshal(inc.Context)
	if err != nil {
		return inc, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return inc, err
	}
	defer tx.Rollback(ctx)

	saved, err := scanIncident(tx.QueryRow(ctx,
		`INSERT INTO incidents (rule_id, home_id, severity, message, context, triggered_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (rule_id, triggered_at) DO NOTHING
		 RETURNING `+incidentColumns,
		inc.RuleID, inc.HomeID, inc.Severity, inc.Message, contextJSON, inc.TriggeredAt))
	if errors.Is(err, ErrNotFound) {
		// ON CONFLICT skipped the insert: the same rule already fired at this
		// instant (crash replay). The deferred rollback discards the empty
		// transaction; the caller treats this as a duplicate, not a failure.
		return Incident{}, ErrConflict
	}
	if err != nil {
		return inc, fmt.Errorf("inserting incident: %w", err)
	}
	for _, ch := range channels {
		if _, err := tx.Exec(ctx,
			`INSERT INTO notifications (incident_id, channel) VALUES ($1, $2)`, saved.ID, ch); err != nil {
			return inc, fmt.Errorf("inserting notification: %w", err)
		}
	}
	return saved, tx.Commit(ctx)
}

// ListIncidents filters by home and status; empty filters match everything.
// Newest first, at most limit rows.
func (s *Store) ListIncidents(ctx context.Context, homeID, status string, limit int) ([]Incident, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT `+incidentColumns+` FROM incidents
		 WHERE ($1 = '' OR home_id::text = $1) AND ($2 = '' OR status = $2)
		 ORDER BY triggered_at DESC
		 LIMIT $3`, homeID, status, limit)
	if err != nil {
		return nil, fmt.Errorf("listing incidents: %w", err)
	}
	defer rows.Close()

	incidents := []Incident{}
	for rows.Next() {
		inc, err := scanIncident(rows)
		if err != nil {
			return nil, err
		}
		incidents = append(incidents, inc)
	}
	return incidents, rows.Err()
}

// AcknowledgeIncident moves an open incident to acknowledged.
func (s *Store) AcknowledgeIncident(ctx context.Context, id string) (Incident, error) {
	return s.transition(ctx, id,
		`UPDATE incidents SET status = 'acknowledged', acknowledged_at = now()
		 WHERE id = $1 AND status = 'open' RETURNING `+incidentColumns)
}

// ResolveIncident closes an open or acknowledged incident.
func (s *Store) ResolveIncident(ctx context.Context, id string) (Incident, error) {
	return s.transition(ctx, id,
		`UPDATE incidents SET status = 'resolved', resolved_at = now(),
		     acknowledged_at = coalesce(acknowledged_at, now())
		 WHERE id = $1 AND status <> 'resolved' RETURNING `+incidentColumns)
}

func (s *Store) transition(ctx context.Context, id, query string) (Incident, error) {
	inc, err := scanIncident(s.pool.QueryRow(ctx, query, id))
	if !errors.Is(err, ErrNotFound) {
		return inc, err
	}
	// No row updated: either the incident does not exist or it is already
	// past this state.
	if _, getErr := scanIncident(s.pool.QueryRow(ctx,
		`SELECT `+incidentColumns+` FROM incidents WHERE id = $1`, id)); getErr != nil {
		return inc, getErr
	}
	return inc, fmt.Errorf("%w: incident is already past that state", ErrConflict)
}
