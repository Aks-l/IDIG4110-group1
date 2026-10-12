package simulator

import (
	"testing"
	"time"

	"IDIG4110/device-simulator/internal/config"
	"IDIG4110/shared/dto"
)

func testHouse() config.House {
	return config.House{
		ID: "h1", Name: "Home", GatewayID: "g1",
		Rooms: []config.Room{{ID: "r1", Name: "Room", Devices: []config.Device{
			{ID: "light_1", Type: "smart_light", InitialState: map[string]any{"on": false, "brightness": 50.0}},
			{ID: "temp_1", Type: "temperature_sensor", InitialState: map[string]any{"temperature": 21.0}},
		}}},
	}
}

func TestGatewayBuildsConfiguredDevicesAndSimulates(t *testing.T) {
	gateway, err := NewGateway(testHouse(), NewMemoryStorage())
	if err != nil {
		t.Fatal(err)
	}
	if len(gateway.Devices()) != 2 {
		t.Fatalf("expected two devices, got %d", len(gateway.Devices()))
	}
	readings, err := gateway.Simulate(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(readings) == 0 {
		t.Fatal("expected simulated readings")
	}
}

func TestGatewayRejectsUnsupportedControl(t *testing.T) {
	gateway, err := NewGateway(testHouse(), NewMemoryStorage())
	if err != nil {
		t.Fatal(err)
	}
	_, err = gateway.ApplyCommand(dto.Command{ExternalEntityID: "temp_1", Command: "turn_on"}, time.Now())
	if err == nil {
		t.Fatal("expected read-only device command to fail")
	}
}

func TestGatewayUpdatesSupportedDevice(t *testing.T) {
	gateway, err := NewGateway(testHouse(), NewMemoryStorage())
	if err != nil {
		t.Fatal(err)
	}
	readings, err := gateway.ApplyCommand(dto.Command{
		ExternalEntityID: "light_1", Command: "turn_on",
	}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(readings) == 0 {
		t.Fatal("expected state reading after command")
	}
}

func TestGatewayAppliesEntityAddressedCommand(t *testing.T) {
	gateway, err := NewGateway(testHouse(), NewMemoryStorage())
	if err != nil {
		t.Fatal(err)
	}
	readings, err := gateway.ApplyCommand(dto.Command{
		ExternalEntityID: "light_1.on", Command: "turn_on",
	}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	state, ok := gateway.DeviceState("light_1")
	if !ok || state["on"] != true {
		t.Fatalf("expected light_1 to be on after command, got %v", state)
	}
	for _, reading := range readings {
		if reading.DeviceID == "light_1" && reading.Property == "on" && reading.State == "true" {
			return
		}
	}
	t.Fatal("expected on reading after entity addressed command")
}

func TestGatewayRejectsUnknownEntityCommand(t *testing.T) {
	gateway, err := NewGateway(testHouse(), NewMemoryStorage())
	if err != nil {
		t.Fatal(err)
	}
	_, err = gateway.ApplyCommand(dto.Command{ExternalEntityID: "unknown_1.on", Command: "turn_on"}, time.Now())
	if err == nil {
		t.Fatal("expected unknown entity command to fail")
	}
}

func TestConfiguredMotionProbabilityCanForceMotion(t *testing.T) {
	always := 1.0
	house := testHouse()
	house.Rooms[0].Devices = append(house.Rooms[0].Devices, config.Device{
		ID: "motion_1", Type: "motion_sensor",
		Simulation: config.SimulationConfig{MotionProbability: &always},
	})
	gateway, err := NewGateway(house, NewMemoryStorage())
	if err != nil {
		t.Fatal(err)
	}
	readings, err := gateway.Simulate(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, reading := range readings {
		if reading.DeviceID == "motion_1" && reading.Property == "motion" && reading.State == "true" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected configured motion sensor to report true")
	}
}
