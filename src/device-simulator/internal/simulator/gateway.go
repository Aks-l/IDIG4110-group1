package simulator

import (
	"fmt"
	"sync"
	"time"

	"IDIG4110/device-simulator/internal/config"
	"IDIG4110/shared/dto"
)

type Gateway struct {
	mu      sync.RWMutex
	id      string
	houseID string
	house   HouseInfo
	devices map[string]Device
	storage Storage
}

type HouseInfo struct {
	ID        string
	Name      string
	GatewayID string
	Rooms     []RoomInfo
}

type RoomInfo struct {
	ID   string
	Name string
}

func NewGateway(house config.House, storage Storage) (*Gateway, error) {
	if storage == nil {
		storage = NewMemoryStorage()
	}
	gateway := &Gateway{
		id: house.GatewayID, houseID: house.ID,
		house:   HouseInfo{ID: house.ID, Name: house.Name, GatewayID: house.GatewayID},
		devices: map[string]Device{}, storage: storage,
	}
	for _, room := range house.Rooms {
		gateway.house.Rooms = append(gateway.house.Rooms, RoomInfo{ID: room.ID, Name: room.Name})
		for _, deviceConfig := range room.Devices {
			device, err := NewDevice(room.ID, deviceConfig)
			if err != nil {
				return nil, err
			}
			if _, exists := gateway.devices[deviceConfig.ID]; exists {
				return nil, fmt.Errorf("duplicate device id %q", deviceConfig.ID)
			}
			gateway.devices[deviceConfig.ID] = device
		}
	}
	return gateway, nil
}

func (g *Gateway) ID() string {
	return g.id
}

func (g *Gateway) Devices() []DeviceInfo {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make([]DeviceInfo, 0, len(g.devices))
	for _, device := range g.devices {
		result = append(result, device.Info())
	}
	return result
}

func (g *Gateway) House() HouseInfo {
	g.mu.RLock()
	defer g.mu.RUnlock()
	house := g.house
	house.Rooms = append([]RoomInfo(nil), g.house.Rooms...)
	return house
}

func (g *Gateway) DeviceState(deviceID string) (map[string]any, bool) {
	g.mu.RLock()
	device, ok := g.devices[deviceID]
	g.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return device.State(), true
}

func (g *Gateway) Simulate(now time.Time) ([]Reading, error) {
	g.mu.RLock()
	devices := make([]Device, 0, len(g.devices))
	for _, device := range g.devices {
		devices = append(devices, device)
	}
	g.mu.RUnlock()

	var readings []Reading
	for _, device := range devices {
		for _, reading := range device.Simulate(now) {
			if err := g.storage.SaveReading(reading); err != nil {
				return nil, fmt.Errorf("storing reading for %q: %w", reading.DeviceID, err)
			}
			readings = append(readings, reading)
		}
	}
	return readings, nil
}

func (g *Gateway) ApplyCommand(command dto.Command, now time.Time) ([]Reading, error) {
	g.mu.RLock()
	device, ok := g.devices[command.ExternalEntityID]
	g.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("device %q not found on gateway %q", command.ExternalEntityID, g.id)
	}
	if err := device.ApplyCommand(command); err != nil {
		return nil, err
	}
	if err := g.storage.SaveStateChange(StateChange{DeviceID: command.ExternalEntityID, State: device.State(), Time: now}); err != nil {
		return nil, err
	}
	return device.Simulate(now), nil
}

func (g *Gateway) Storage() Storage {
	return g.storage
}
