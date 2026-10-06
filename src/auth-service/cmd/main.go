package main

import (
	"IDIG4110/auth-service/internal/server"
	"log/slog"
)

func main() {
	if err := server.Run(); err != nil {
		slog.Error("Starting server", "error", err)
	}
}
