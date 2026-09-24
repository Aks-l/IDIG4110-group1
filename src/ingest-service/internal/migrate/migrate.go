package migrate

import (
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
)

type Migrator struct {
	Migrate *migrate.Migrate
}

func Init(url string) (*Migrator, error) {
	m, err := migrate.New(
		"file:///migrations",
		url,
	)
	if err != nil {
		return nil, err
	}
	slog.Info("Initilizing database migrator")

	return &Migrator{
		Migrate: m,
	}, nil
}

func (m *Migrator) CheckMigrationStatus() error {
	version, dirty, error := m.Migrate.Version()
	if error != nil {
		return error
	}
	if dirty {
		return fmt.Errorf("database in dirty state at version: %d", version)
	}

	slog.Info("Database schema", "version", version)
	return nil
}
