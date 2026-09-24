package db

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"IDIG4110/ingest-service/internal/config"

	"github.com/jackc/pgx/v5"
)

type Database struct {
	conn *pgx.Conn
}

func Init(cfg config.DatabaseConfig) (Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.User, cfg.Password, cfg.Host, strconv.Itoa(cfg.Port), cfg.Name)

	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return Database{}, err
	}

	if err := conn.Ping(ctx); err != nil {
		return Database{}, err
	}

	slog.Info("Successfully connected to database")

	return Database{
		conn: conn,
	}, nil
}

func (d *Database) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	d.conn.Close(ctx)
}
