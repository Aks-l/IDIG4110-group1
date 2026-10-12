package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadHouse(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "house.json")
	content := `{"house":{"id":"h1","name":"Home","gateway_id":"g1","rooms":[{"id":"r1","name":"Room","devices":[{"id":"d1","type":"temperature_sensor","initial_state":{"temperature":21.5}}]}]}}`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	house, err := LoadHouse(path)
	if err != nil {
		t.Fatal(err)
	}
	if house.House.Rooms[0].Devices[0].ID != "d1" {
		t.Fatalf("unexpected device: %+v", house.House.Rooms[0].Devices[0])
	}
}

func TestLoadHouseRejectsDuplicateDeviceIDs(t *testing.T) {
	cfg := HouseConfig{House: House{
		ID: "h1", Name: "Home", GatewayID: "g1",
		Rooms: []Room{{ID: "r1", Name: "Room", Devices: []Device{{ID: "d1", Type: "temperature_sensor"}, {ID: "d1", Type: "humidity_sensor"}}}},
	}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected duplicate device error")
	}
}

func TestLoadHouseRejectsInvalidSimulationSettings(t *testing.T) {
	probability := 2.0
	cfg := HouseConfig{House: House{
		ID: "h1", Name: "Home", GatewayID: "g1",
		Rooms: []Room{{ID: "r1", Name: "Room", Devices: []Device{{
			ID: "motion_1", Type: "motion_sensor",
			Simulation: SimulationConfig{MotionProbability: &probability},
		}}}},
	}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid probability error")
	}
}
