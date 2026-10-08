package db

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"IDIG4110/shared/dto"
	"IDIG4110/twin-core/internal/config"
	"IDIG4110/twin-core/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Conn *pgxpool.Pool
	Url  string
}

func Init(cfg config.DatabaseConfig) (*Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&search_path=public", cfg.User, cfg.Password, cfg.Host, strconv.Itoa(cfg.Port), cfg.Name, cfg.SSL)

	poolCfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return &Database{}, err
	}

	// Superusers bypass row level security and the postgres image makes
	// the bootstrap user one (PG17 even refuses to demote it). The first
	// migration creates the no-login runtime role; every connection drops
	// into it so the home_isolation policies bind. Until the role exists
	// (fresh database, migrations not applied yet) the session stays as-is.
	poolCfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SELECT set_config('role', 'twin_app', false) WHERE EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'twin_app')")
		return err
	}

	conn, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return &Database{}, err
	}

	if err := conn.Ping(ctx); err != nil {
		return &Database{}, err
	}

	slog.Info("Successfully connected to database")

	return &Database{
		Conn: conn,
		Url:  url,
	}, nil
}

func (d *Database) Close() {
	d.Conn.Close()
}

// Runs fn in a transaction scoped to one home
// row level security policies on the home tables read app.home_id,
// set_config with local=true keeps the value inside the transaction so it
// cannot leak to the next pool user; an unset app.home_id leaves the
// policies hiding every row
//
// # Inputs:
//
//   - ctx [context.Context] request context
//   - homeID [string] home id
//   - fn [func(pgx.Tx) error] queries to run under the home context
//
// # Returns:
//
//   - Error from fn, commit or setup
func (d *Database) WithHome(ctx context.Context, homeID string, fn func(pgx.Tx) error) error {
	if !dto.IsValidUUID(homeID) {
		return fmt.Errorf("%w: home id %q is not a uuid", domain.ErrBadRequest, homeID)
	}

	tx, err := d.Conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin home transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_config('app.home_id', $1, true)", homeID); err != nil {
		return fmt.Errorf("set home context: %w", err)
	}

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
