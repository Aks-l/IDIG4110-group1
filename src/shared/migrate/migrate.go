// Package migrate wraps golang-migrate for use by any service in this repo.
// It is config-agnostic: callers pass the migrations directory and database URL.
package migrate

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Migrator struct {
	Migrate *migrate.Migrate
}

func Init(directory, url string) (*Migrator, error) {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", directory),
		url,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initilze migrator: %w", err)
	}
	slog.Info("Initilizing database migrator")

	return &Migrator{
		Migrate: m,
	}, nil
}

// Up applies all pending migrations. Having nothing to apply is not an error.
func (m *Migrator) Up() error {
	if err := m.Migrate.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("applying migrations: %w", err)
	}
	return nil
}

func (m *Migrator) CheckMigrationStatus() error {
	version, dirty, err := m.Migrate.Version()
	if err != nil {
		return fmt.Errorf("check migration status: %w", err)
	}
	if dirty {
		return fmt.Errorf("check migration status: database in dirty state at version: %d", version)
	}

	slog.Info("Database schema", "version", version)
	return nil
}
