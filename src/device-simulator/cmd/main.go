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

	house, err := config.LoadHouse(cfg.HouseConfig)
	if err != nil {
		slog.Error("house config failed", "error", err)
		os.Exit(1)
	}
	s, err := simulator.Init(b, cfg.Mqtt, house.House)
	if err != nil {
		slog.Error("simulator failed", "error", err)
		os.Exit(1)
	}
	if err := s.Start(); err != nil {
		slog.Error("simulator stopped", "error", err)
		os.Exit(1)
	}
}
