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

	"IDIG4110/shared/migrate"
	"IDIG4110/twin-core/internal/config"
	"IDIG4110/twin-core/internal/db"
	"IDIG4110/twin-core/internal/repository"
	"IDIG4110/twin-core/internal/service"
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
	if err := m.Up(); err != nil {
		return err
	}
	if err := m.CheckMigrationStatus(); err != nil {
		return err
	}

	stateRepo := repository.NewTwinStateRepoImpl(db)
	twinStateSvc := service.NewImplTwinStateSvc(stateRepo)
	stateQuerySvc := service.NewImplStateQuerySvc(stateRepo)
	structureSvc := service.NewImplStructureSvc(stateRepo)

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

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverError:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen: %w", err)
		}
		return nil
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
