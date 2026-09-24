package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"IDIG4110/ingest-service/internal/config"
)

func Run() error {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		return err
	}

	router := NewRouter()

	httpServer := http.Server{
		Addr:           cfg.Server.Port,
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimout) * time.Second,
		IdleTimeout:    time.Duration(cfg.Server.IdleTimeout) * time.Second,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	serverError := make(chan error, 1)
	go func() {
		slog.Info("Starting server", "port", "8080")
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
