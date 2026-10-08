package main

import (
	"log/slog"

	"IDIG4110/twin-core/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		slog.Error("Starting server", "error", err)
	}
}
