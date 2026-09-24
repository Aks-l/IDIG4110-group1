package main

import (
	"log/slog"

	"IDIG4110/ingest-service/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		slog.Error("Starting server", "error", err)
	}
}
