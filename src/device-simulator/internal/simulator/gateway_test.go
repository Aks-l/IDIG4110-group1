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
