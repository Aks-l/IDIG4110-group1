package server

import (
	"IDIG4110/auth-service/internal/config"
	"IDIG4110/auth-service/internal/db"
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
)

func Run() error {
	cfg := config.Config{}
	cfg.Init()

	db := db.Database{}
	err := db.Init(cfg.Database)
	if err != nil {
		return err
	}

	router := NewRouter()

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
