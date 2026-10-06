package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"IDIG4110/auth-service/internal/config"
	db "IDIG4110/shared/db"
	"IDIG4110/shared/migrate"
)

func main() {
	slog.Info("Staring migration cmd")

	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		os.Exit(1)
	}

	command := args[0]

	cfg := config.Config{}
	cfg.Init()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := db.NewPostgresPool(ctx, cfg.Database.URL)
	if err != nil {
		slog.Error("Failed to init db", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	m, err := migrate.Init(cfg.Migration.Directory, cfg.Database.URL)
	if err != nil {
		slog.Error("Failed to init db", "error", err)
		os.Exit(1)
	}

	switch command {
	case "up":
		handleUp(ctx, m, args)
	case "down":
		handleDown(ctx, m, args)
	case "version":
		handleVersion(m)
	case "create":
		handleCreate(cfg.Migration.Directory, args)
	default:
		slog.Error("Invalid command", "command", command)
		os.Exit(1)
	}
}

func handleUp(ctx context.Context, m *migrate.Migrator, args []string) {
	if len(args) > 1 {
		n, err := strconv.Atoi(args[1])
		if err != nil || n < 1 {
			slog.Error("Invalid number of steps", "value", args[1])
			os.Exit(1)
		}
		if err := m.Migrate.Steps(n); err != nil {
			slog.Error("Migration error", "err", err)
			os.Exit(1)
		}
	} else {
		if err := m.Migrate.Up(); err != nil {
			slog.Error("Migration error", "err", err)
			os.Exit(1)
		}
	}
}

func handleDown(ctx context.Context, m *migrate.Migrator, args []string) {
	steps := 1
	if len(args) > 1 {
		n, err := strconv.Atoi(args[1])
		if err != nil || n < 1 {
			slog.Error("Invalid number of steps", "value", args[1])
			os.Exit(1)
		}
		steps = n
	}

	fmt.Printf("WARNING: This will rollback %d migration(s)\n", steps)
	fmt.Print("Type 'yes' to confirm: ")
	var confirmation string
	if _, err := fmt.Scanln(&confirmation); err != nil {
		slog.Error("Failed to read confirmation", "err", err)
		os.Exit(1)
	}
	if confirmation != "yes" {
		slog.Info("Operation cancelled")
		return
	}

	slog.Info("Rolling back migrations...", "steps", steps)
	done := make(chan error, 1)
	go func() {
		done <- m.Migrate.Steps(-steps)
	}()

	select {
	case <-ctx.Done():
		slog.Error("migration timeout exceeded")
		return
	case err := <-done:
		if err != nil {
			slog.Error("rollback failed", "error", err)
			return
		}
		slog.Info("Rollback completed successfully", "status", "ok")
		return
	}
}

func handleCreate(migrationsDir string, args []string) {
	if len(args) < 2 {
		slog.Error("Migration name required")
		fmt.Println("Usage: migrate create NAME")
		os.Exit(1)
	}

	name := args[1]
	timestamp := time.Now().Format("20060102150405")

	upFile := fmt.Sprintf("%s/%s_%s.up.sql", migrationsDir, timestamp, name)
	downFile := fmt.Sprintf("%s/%s_%s.down.sql", migrationsDir, timestamp, name)

	upContent := fmt.Sprintf(`-- Migration: %s
-- Created: %s
-- Description: Add description here

BEGIN;

-- Add your migration SQL here

COMMIT;
`, name, time.Now().Format(time.RFC3339))

	downContent := fmt.Sprintf(`-- Migration: %s (rollback)
-- Created: %s

BEGIN;

-- Add your rollback SQL here

COMMIT;
`, name, time.Now().Format(time.RFC3339))

	if err := os.WriteFile(upFile, []byte(upContent), 0644); err != nil {
		slog.Error("Failed to create up migration", "err", err)
		os.Exit(1)
	}

	if err := os.WriteFile(downFile, []byte(downContent), 0644); err != nil {
		slog.Error("Failed to create down migration", "err", err)
		os.Exit(1)
	}

	slog.Info("Migration files created", "up", upFile, "down", downFile)
}

func handleVersion(m *migrate.Migrator) {
	version, dirty, err := m.Migrate.Version()
	if err != nil {
		slog.Error("Failed to get version", "err", err)
		os.Exit(1)
	}

	fmt.Println("\nMigration Status:")
	fmt.Println("=================")
	fmt.Printf("Current version: %d\n", version)
	if dirty {
		fmt.Println("Status: DIRTY (migration failed or interrupted)")
		fmt.Println("\nTo recover, use: migrate force VERSION")
	} else {
		fmt.Println("Status: Clean")
	}
}
