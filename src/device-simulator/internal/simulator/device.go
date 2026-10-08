package simulator

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"IDIG4110/device-simulator/internal/config"
	"IDIG4110/shared/dto"
)

type Reading struct {
	DeviceID   string
	Property   string
	State      string
	Attributes map[string]string
	Time       time.Time
}

type DeviceInfo struct {
	ID           string
	RoomID       string
	Type         string
	Capabilities []string
}

type Device interface {
	Info() DeviceInfo
	State() map[string]any
	Simulate(time.Time) []Reading
	ApplyCommand(dto.Command) error
}

type genericDevice struct {
	mu          sync.RWMutex
	info        DeviceInfo
	state       map[string]any
	lastReading map[string]string
}

var capabilitiesByType = map[string][]string{
	"temperature_sensor": {"read_temperature"},
	"humidity_sensor":    {"read_humidity"},
	"motion_sensor":      {"read_motion"},
	"smart_light":        {"turn_on", "turn_off", "set_brightness"},
	"air_conditioner":    {"turn_on", "turn_off", "set_temperature", "set_mode", "read_temperature"},
	"smart_oven":         {"turn_on", "turn_off", "set_temperature", "set_mode", "read_temperature"},
	"thermostat":         {"set_temperature", "read_temperature"},
	"door_sensor":        {"read_open"},
	"smart_plug":         {"turn_on", "turn_off", "read_power"},
	"energy_sensor":      {"read_power", "read_energy"},
}

func NewDevice(roomID string, cfg config.Device) (Device, error) {
	capabilities, ok := capabilitiesByType[cfg.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported device type %q", cfg.Type)
	}
	state := make(map[string]any, len(cfg.InitialState))
	for key, value := range cfg.InitialState {
		state[key] = value
	}
	return &genericDevice{
		info:        DeviceInfo{ID: cfg.ID, RoomID: roomID, Type: cfg.Type, Capabilities: append([]string(nil), capabilities...)},
		state:       state,
		lastReading: map[string]string{},
	}, nil
}

func (d *genericDevice) Info() DeviceInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()
	info := d.info
	info.Capabilities = append([]string(nil), d.info.Capabilities...)
	return info
}

func (d *genericDevice) State() map[string]any {
	d.mu.RLock()
	defer d.mu.RUnlock()
	state := make(map[string]any, len(d.state))
	for key, value := range d.state {
		state[key] = value
	}
	return state
}

func (d *genericDevice) Simulate(now time.Time) []Reading {
	d.mu.Lock()
	defer d.mu.Unlock()

	switch d.info.Type {
	case "temperature_sensor":
		d.state["temperature"] = NextTemperature(number(d.state["temperature"], 21))
	case "humidity_sensor":
		d.state["humidity"] = clamp(number(d.state["humidity"], 45)+(rand.Float64()*2-1), 20, 80)
	case "motion_sensor":
		d.state["motion"] = rand.Float64() < 0.15
	case "door_sensor":
		d.state["open"] = rand.Float64() < 0.05
	case "air_conditioner", "smart_oven", "thermostat":
		if d.info.Type != "thermostat" && !boolean(d.state["on"]) {
			break
		}
		current := number(d.state["current_temperature"], 21)
		target := number(d.state["target_temperature"], 21)
		d.state["current_temperature"] = current + clamp(target-current, -0.5, 0.5)
	case "smart_plug":
		if boolean(d.state["on"]) {
			d.state["power"] = 60.0 + rand.Float64()*20
		} else {
			d.state["power"] = 0.0
		}
	case "energy_sensor":
		power := number(d.state["power"], 100)
		d.state["power"] = power + (rand.Float64()*10 - 5)
		d.state["energy"] = number(d.state["energy"], 0) + power/360.0
	}
	return d.readings(now)
}

func (d *genericDevice) readings(now time.Time) []Reading {
	properties := map[string]struct {
		deviceClass string
		unit        string
	}{
		"temperature":         {"temperature", "°C"},
		"humidity":            {"humidity", "%"},
		"motion":              {"motion", ""},
		"open":                {"door", ""},
		"current_temperature": {"temperature", "°C"},
		"target_temperature":  {"target_temperature", "°C"},
		"power":               {"power", "W"},
		"energy":              {"energy", "kWh"},
		"on":                  {"switch", ""},
		"brightness":          {"brightness", "%"},
		"mode":                {"mode", ""},
	}
	var result []Reading
	for property, metadata := range properties {
		value, ok := d.state[property]
		if !ok {
			continue
		}
		state := formatValue(value)
		if d.lastReading[property] == state && property != "on" && property != "mode" {
			continue
		}
		d.lastReading[property] = state
		attributes := map[string]string{"device_class": metadata.deviceClass}
		if metadata.unit != "" {
			attributes["unit_of_measurement"] = metadata.unit
		}
		result = append(result, Reading{DeviceID: d.info.ID, Property: property, State: state, Attributes: attributes, Time: now})
	}
	return result
}

func (d *genericDevice) ApplyCommand(command dto.Command) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if command.ExternalEntityID != d.info.ID && command.ExternalEntityID != d.topicID() {
		return fmt.Errorf("command targets %q, device is %q", command.ExternalEntityID, d.info.ID)
	}
	switch command.Command {
	case "turn_on", "turn_off":
		if !hasCapability(d.info.Capabilities, command.Command) {
			return fmt.Errorf("device %q does not support %s", d.info.ID, command.Command)
		}
		d.state["on"] = command.Command == "turn_on"
	case "set_brightness":
		if !hasCapability(d.info.Capabilities, command.Command) {
			return fmt.Errorf("device %q does not support set_brightness", d.info.ID)
		}
		brightness, err := commandNumber(command.Parameters, "brightness")
		if err != nil || brightness < 0 || brightness > 100 {
			return fmt.Errorf("brightness must be a number between 0 and 100")
		}
		d.state["brightness"] = brightness
	case "set_temperature":
		if !hasCapability(d.info.Capabilities, command.Command) {
			return fmt.Errorf("device %q does not support set_temperature", d.info.ID)
		}
		temperature, err := commandNumber(command.Parameters, "temperature")
		if err != nil {
			return fmt.Errorf("temperature must be a number")
		}
		d.state["target_temperature"] = temperature
	case "set_mode":
		if !hasCapability(d.info.Capabilities, command.Command) {
			return fmt.Errorf("device %q does not support set_mode", d.info.ID)
		}
		mode, ok := command.Parameters["mode"].(string)
		if !ok || mode == "" {
			return fmt.Errorf("set_mode requires a string mode")
		}
		d.state["mode"] = mode
	default:
		return fmt.Errorf("unsupported command %q", command.Command)
	}
	return nil
}

func (d *genericDevice) topicID() string {
	return "sensors/" + d.info.ID + "/state"
}

func hasCapability(capabilities []string, wanted string) bool {
	for _, capability := range capabilities {
		if capability == wanted {
			return true
		}
	}
	return false
}

func commandNumber(parameters map[string]any, key string) (float64, error) {
	return numberWithError(parameters[key])
}

func numberWithError(value any) (float64, error) {
	switch value := value.(type) {
	case float64:
		return value, nil
	case float32:
		return float64(value), nil
	case int:
		return float64(value), nil
	default:
		return 0, fmt.Errorf("value is not numeric")
	}
}

func number(value any, fallback float64) float64 {
	switch value := value.(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	default:
		return fallback
	}
}

func boolean(value any) bool {
	booleanValue, ok := value.(bool)
	return ok && booleanValue
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func formatValue(value any) string {
	switch value := value.(type) {
	case float64:
		return strconv.FormatFloat(value, 'f', 1, 64)
	case float32:
		return strconv.FormatFloat(float64(value), 'f', 1, 64)
	case bool:
		return strconv.FormatBool(value)
	default:
		return fmt.Sprint(value)
	}
}
