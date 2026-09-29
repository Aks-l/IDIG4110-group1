package migrate

import (
	"fmt"
	"log/slog"

	"IDIG4110/ingest-service/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Migrator struct {
	Migrate *migrate.Migrate
}

func Init(cfg config.MigrationConfig, url string) (*Migrator, error) {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", cfg.Directory),
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
