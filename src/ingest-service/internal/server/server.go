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

	"IDIG4110/ingest-service/internal/config"
	"IDIG4110/ingest-service/internal/db"
	"IDIG4110/shared/migrate"
	"IDIG4110/ingest-service/internal/mqttclient"
	"IDIG4110/ingest-service/internal/repository"
	"IDIG4110/ingest-service/internal/service"
)

func Run() error {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		return err
	}

	db, err := db.Init(cfg.Database)
	if err != nil {
		return err
	}
	defer db.Close()
	m, err := migrate.Init(cfg.Migration.Directory, db.Url)
	if err != nil {
		return err
	}
	m.Migrate.Up()

	if err := m.CheckMigrationStatus(); err != nil {
		return err
	}

	sensorIngestRepo := repository.NewSensorIngestRepoImpl(db)
	sensorIngestSvc := service.NewImplSensorIngestSvc(sensorIngestRepo)

	coll := mqttclient.NewCollector(
		cfg.Mqtt.WorkerCount,
		cfg.Mqtt.WorkerBufferSize,
		sensorIngestSvc,
	)
	coll.StartWorkers()
	defer coll.Close()

	client, err := mqttclient.Init(cfg.Mqtt, coll.MQTTHandler)
	if err != nil {
		return err
	}
	slog.Info("starting mqtt client", "topic", cfg.Mqtt.Topic)

	defer client.Close()

	if err := client.Subscribe(cfg.Mqtt.Topic); err != nil {
		return err
	}

	router := NewRouter(sensorIngestSvc)

	httpServer := http.Server{
		Addr:           ":" + strconv.Itoa(cfg.Server.Port),
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimout) * time.Second,
		IdleTimeout:    time.Duration(cfg.Server.IdleTimeout) * time.Second,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	serverError := make(chan error, 1)
	go func() {
		slog.Info("Starting server", "port", strconv.Itoa(cfg.Server.Port))
		err := httpServer.ListenAndServe()
		serverError <- err
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverError:
		if err != nil {
			return nil
		}
		if errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return fmt.Errorf("listen %w", err)
	case <-shutdown:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		slog.Info("Closing server")

		if err := httpServer.Shutdown(ctx); err != nil {
			return err
		}
	}

	return nil
}
