// Command rules-engine evaluates automation rules against the reading stream
// and serves the rules and incidents API. See docs/architecture/rules-engine.md.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"IDIG4110/rules-engine/internal/api"
	"IDIG4110/rules-engine/internal/config"
	"IDIG4110/rules-engine/internal/engine"
	"IDIG4110/rules-engine/internal/executor"
	"IDIG4110/rules-engine/internal/runner"
	"IDIG4110/rules-engine/internal/store"
	"IDIG4110/shared/kafka"
	"IDIG4110/shared/migrate"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		slog.Error("rules-engine stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	m, err := migrate.Init(cfg.MigrationsDir, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil {
		return err
	}
	if err := m.CheckMigrationStatus(); err != nil {
		return err
	}

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("database: %w", err)
	}

	st := store.New(pool)
	eng := engine.New()
	reload := func(ctx context.Context) error {
		rules, err := st.ListRules(ctx, "")
		if err != nil {
			return err
		}
		eng.SetRules(rules)
		return nil
	}
	if err := reload(ctx); err != nil {
		return fmt.Errorf("loading rules: %w", err)
	}

	producer, err := kafka.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		return err
	}
	defer producer.Close()
	consumer, err := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaGroup, kafka.TopicReadings)
	if err != nil {
		return err
	}
	defer consumer.Close()

	exec := executor.New(st, producer, cfg.CommandTTL)
	go func() {
		if err := runner.New(consumer, eng, exec, reload, cfg.TickInterval, cfg.RefreshInterval).Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("Reading stream stopped", "error", err)
			stop()
		}
	}()

	server := &http.Server{
		Addr:              ":" + strconv.Itoa(cfg.HTTPPort),
		Handler:           api.New(st, reload).Routes(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("rules-engine listening", "port", cfg.HTTPPort, "brokers", cfg.KafkaBrokers)
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server: %w", err)
		}
	case <-ctx.Done():
		slog.Info("Shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}
