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
	Conn *pgx.Conn
	Url string
}

func Init(cfg config.DatabaseConfig) (*Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", cfg.User, cfg.Password, cfg.Host, strconv.Itoa(cfg.Port), cfg.Name, cfg.SSL)

	conn, err := pgx.Connect(ctx, url)
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	d.Conn.Close(ctx)
}
