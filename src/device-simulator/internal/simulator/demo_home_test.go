package simulator

import (
	"testing"
	"time"

	"IDIG4110/device-simulator/internal/config"
	"IDIG4110/shared/dto"
)

// demoHouse mirrors the parts of the seeded demo home these tests need: a
// climate device with two configured entity ids, a lock, a light whose
// entity id differs from its device id, a valve and a binary sensor.
func demoHouse() config.House {
	return config.House{
		ID: "h1", Name: "Demo", GatewayID: "g1",
		Rooms: []config.Room{{ID: "r1", Name: "Room", Devices: []config.Device{
			{
				ID:           "sensor.bedroom_climate",
				Type:         "climate_sensor",
				InitialState: map[string]any{"temperature": 20.0, "humidity": 41.0},
				Entities:     map[string]string{"temperature": "temperature.bedroom", "humidity": "humidity.bedroom"},
			},
			{
				ID:           "lock.front_door",
				Type:         "lock",
				InitialState: map[string]any{"locked": "locked"},
				Entities:     map[string]string{"locked": "lock.front_door"},
			},
			{
				ID:           "light.hall",
				Type:         "light",
				InitialState: map[string]any{"on": "off"},
				Entities:     map[string]string{"on": "light.hall_main"},
			},
			{
				ID:           "valve.main_water",
				Type:         "valve",
				InitialState: map[string]any{"open": "open"},
				Entities:     map[string]string{"open": "valve.main_water"},
			},
			{
				ID:           "binary_sensor.kitchen_smoke",
				Type:         "binary_sensor",
				InitialState: map[string]any{"smoke": "off"},
				Entities:     map[string]string{"smoke": "binary_sensor.kitchen_smoke"},
			},
		}}},
	}
}

func TestConfiguredEntityIDsOverrideComposition(t *testing.T) {
	gateway, err := NewGateway(demoHouse(), NewMemoryStorage())
	if err != nil {
		t.Fatal(err)
	}
	readings, err := gateway.Simulate(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]string{}
	for _, reading := range readings {
		ids[reading.Property] = reading.EntityID
	}
	if ids["temperature"] != "temperature.bedroom" || ids["humidity"] != "humidity.bedroom" {
		t.Fatalf("expected configured entity ids, got %v", ids)
	}
	if ids["locked"] != "lock.front_door" {
		t.Fatalf("expected lock entity id, got %v", ids)
	}
}

func TestUnmappedPropertyFallsBackToComposition(t *testing.T) {
	house := demoHouse()
	house.Rooms[0].Devices = append(house.Rooms[0].Devices, config.Device{
		ID:           "unmapped_1",
		Type:         "temperature_sensor",
		InitialState: map[string]any{"temperature": 21.0},
	})
	gateway, err := NewGateway(house, NewMemoryStorage())
	if err != nil {
		t.Fatal(err)
	}
	readings, err := gateway.Simulate(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, reading := range readings {
		if reading.DeviceID == "unmapped_1" && reading.EntityID != "unmapped_1.temperature" {
			t.Fatalf("expected fallback entity id unmapped_1.temperature, got %q", reading.EntityID)
		}
	}
}

func TestLockCommandUpdatesStateAndPublishes(t *testing.T) {
	gateway, err := NewGateway(demoHouse(), NewMemoryStorage())
	if err != nil {
		t.Fatal(err)
	}
	readings, err := gateway.ApplyCommand(dto.Command{
		ExternalEntityID: "lock.front_door", Command: "unlock",
	}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	state, ok := gateway.DeviceState("lock.front_door")
	if !ok || state["locked"] != "unlocked" {
		t.Fatalf("expected lock to be unlocked, got %v", state)
	}
	for _, reading := range readings {
		if reading.EntityID == "lock.front_door" && reading.State == "unlocked" {
			return
		}
	}
	t.Fatal("expected unlocked reading for lock.front_door")
}

func TestLightReportsOnOffText(t *testing.T) {
	gateway, err := NewGateway(demoHouse(), NewMemoryStorage())
	if err != nil {
		t.Fatal(err)
	}
	readings, err := gateway.ApplyCommand(dto.Command{
		ExternalEntityID: "light.hall_main", Command: "turn_on",
	}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	state, ok := gateway.DeviceState("light.hall")
	if !ok || state["on"] != "on" {
		t.Fatalf("expected light to report on, got %v", state)
	}
	for _, reading := range readings {
		if reading.EntityID == "light.hall_main" && reading.State == "on" {
			return
		}
	}
	t.Fatal("expected on reading for light.hall_main")
}

func TestValveCloseUpdatesState(t *testing.T) {
	gateway, err := NewGateway(demoHouse(), NewMemoryStorage())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := gateway.ApplyCommand(dto.Command{
		ExternalEntityID: "valve.main_water", Command: "close",
	}, time.Now()); err != nil {
		t.Fatal(err)
	}
	state, _ := gateway.DeviceState("valve.main_water")
	if state["open"] != "closed" {
		t.Fatalf("expected valve closed, got %v", state)
	}
}

func TestConfiguredStateProbabilityCanForceBinarySensor(t *testing.T) {
	always := 1.0
	house := demoHouse()
	for i := range house.Rooms[0].Devices {
		if house.Rooms[0].Devices[i].Type == "binary_sensor" {
			house.Rooms[0].Devices[i].Simulation.StateProbability = &always
		}
	}
	gateway, err := NewGateway(house, NewMemoryStorage())
	if err != nil {
		t.Fatal(err)
	}
	readings, err := gateway.Simulate(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, reading := range readings {
		if reading.EntityID == "binary_sensor.kitchen_smoke" && reading.State == "on" {
			return
		}
	}
	t.Fatal("expected forced binary sensor to report on")
}
