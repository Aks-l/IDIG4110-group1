package main

import (
	"IDIG4110/device-simulator/internal/broker"
	"IDIG4110/device-simulator/internal/config"
	"IDIG4110/device-simulator/internal/simulator"
	"log/slog"
	"os"
)


func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		slog.Error("config failed", "error", err)
		os.Exit(1)
	}
	b, err := broker.Init(cfg.Mqtt)
	if err != nil {
		slog.Error("broker failed", "error", err)
		os.Exit(1)	
	}
	defer b.Close()

	s := simulator.Init(b, cfg.Mqtt)
	s.Start()
}
