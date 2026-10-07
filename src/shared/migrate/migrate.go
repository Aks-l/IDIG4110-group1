// Package migrate: golang-migrate wrapper for any service
// directory and database url passed by caller
package migrate

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Thin wrapper around golang-migrate
type Migrator struct {
	Migrate *migrate.Migrate
}

// Creates migrator from migrations directory and database url
//
// # Inputs:
//
//   - directory [string] path to migration files
//   - url [string] postgres connection url
//
// # Returns:
//
//   - Migrator, initialized migrator instance
//   - Init error if initialization fails
func Init(directory, url string) (*Migrator, error) {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", directory),
		url,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migrator: %w", err)
	}
	slog.Info("Initilizing database migrator")

	return &Migrator{
		Migrate: m,
	}, nil
}

// Logs current schema version
//
// # Returns:
//
//   - Error on version check failure or dirty state
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

// Applies all pending migrations
//
// # Returns:
//
//   - Error on migration failure, ErrNoChange ignored
func (m *Migrator) Up() error {
	if err := m.Migrate.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}
