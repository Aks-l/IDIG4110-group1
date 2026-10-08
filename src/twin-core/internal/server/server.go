package server

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

	"IDIG4110/shared/kafka"
	"IDIG4110/shared/migrate"
	"IDIG4110/twin-core/internal/config"
	"IDIG4110/twin-core/internal/consumer"
	"IDIG4110/twin-core/internal/db"
	"IDIG4110/twin-core/internal/repository"
	"IDIG4110/twin-core/internal/service"
)

func Run() error {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		return err
	}

	database, err := db.Init(cfg.Database)
	if err != nil {
		return err
	}
	defer func() { database.Close() }()

	m, err := migrate.Init(cfg.Migration.Directory, database.Url)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil {
		return err
	}
	if err := m.CheckMigrationStatus(); err != nil {
		return err
	}

	// A fresh database just created the runtime role the pool drops into
	// (see db.Init); re-create the pool so no pre-migration connection
	// survives as a row level security bypassing superuser
	database.Close()
	database, err = db.Init(cfg.Database)
	if err != nil {
		return err
	}

	stateRepo := repository.NewTwinStateRepoImpl(database)
	twinStateSvc := service.NewImplTwinStateSvc(stateRepo)
	stateQuerySvc := service.NewImplStateQuerySvc(stateRepo)
	structureSvc := service.NewImplStructureSvc(stateRepo)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Keep twin_state in sync with the readings stream on the event bus
	// (docs/decisions/0001-kafka-event-bus.md). The shared consumer logs
	// and skips handler errors: liveness over completeness.
	if brokers := kafka.ParseBrokers(cfg.Kafka.Brokers); len(brokers) > 0 {
		readingsConsumer, err := kafka.NewConsumer(brokers, cfg.Kafka.Group, kafka.TopicReadings)
		if err != nil {
			return err
		}
		defer readingsConsumer.Close()
		go func() {
			slog.Info("Consuming readings", "topic", kafka.TopicReadings, "group", cfg.Kafka.Group)
			if err := readingsConsumer.Run(ctx, consumer.NewReadings(twinStateSvc).Handle); err != nil && !errors.Is(err, context.Canceled) {
				slog.Error("Readings stream stopped", "error", err)
				stop()
			}
		}()
	} else {
		slog.Warn("No kafka brokers configured, twin_state only changes through POST /api/v1/readings")
	}

	router := NewRouter(stateQuerySvc, twinStateSvc, structureSvc)

	httpServer := http.Server{
		Addr:           ":" + strconv.Itoa(cfg.Server.Port),
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:    time.Duration(cfg.Server.IdleTimeout) * time.Second,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	serverError := make(chan error, 1)
	go func() {
		slog.Info("Starting server", "port", strconv.Itoa(cfg.Server.Port))
		serverError <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverError:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen: %w", err)
		}
	case <-ctx.Done():
		slog.Info("Closing server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}
	}

	return nil
}
