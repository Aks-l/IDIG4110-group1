package main

import (
	"log/slog"

	"IDIG4110/api-gateway/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		slog.Error("Starting server", "error", err)
	}
}
