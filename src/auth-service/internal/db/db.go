package db

import (
	"IDIG4110/auth-service/internal/config"
	database "IDIG4110/shared/db"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool *pgxpool.Pool
}

func (db *Database) Init(cfg config.DatabaseConfig) error {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db.Pool, err = database.NewPostgresPool(ctx, cfg.URL)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) Close() {
	db.Pool.Close()
}
