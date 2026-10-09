package simulator

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"IDIG4110/device-simulator/internal/config"
	"IDIG4110/shared/dto"
)

type Reading struct {
	DeviceID   string
	Property   string
	EntityID   string
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
	Addresses(externalEntityID string) bool
}

type genericDevice struct {
	mu          sync.RWMutex
	info        DeviceInfo
	state       map[string]any
	lastReading map[string]string
	simulation  config.SimulationConfig
	entities    map[string]string
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
	"climate_sensor":     {"read_temperature", "read_humidity"},
	"air_quality_sensor": {"read_humidity", "read_co2"},
	"light":              {"turn_on", "turn_off"},
	"fan":                {"turn_on", "turn_off"},
	"siren":              {"turn_on", "turn_off"},
	"lock":               {"lock", "unlock"},
	"valve":              {"open", "close"},
	"binary_sensor":      {},
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
	entities := make(map[string]string, len(cfg.Entities))
	for property, entityID := range cfg.Entities {
		entities[property] = entityID
	}
	return &genericDevice{
		info:        DeviceInfo{ID: cfg.ID, RoomID: roomID, Type: cfg.Type, Capabilities: append([]string(nil), capabilities...)},
		state:       state,
		lastReading: map[string]string{},
		simulation:  cfg.Simulation,
		entities:    entities,
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

	if d.simulation.Enabled != nil && !*d.simulation.Enabled {
		return d.readings(now)
	}

	switch d.info.Type {
	case "temperature_sensor":
		d.state["temperature"] = d.nextTemperature(number(d.state["temperature"], 21))
	case "climate_sensor":
		d.state["temperature"] = d.nextTemperature(number(d.state["temperature"], 21))
		humidity := d.simulationRange(d.simulation.Humidity, 20, 80, 1)
		d.state["humidity"] = clamp(number(d.state["humidity"], 45)+(rand.Float64()*2-1)*humidity.change, humidity.min, humidity.max)
	case "air_quality_sensor":
		humidity := d.simulationRange(d.simulation.Humidity, 20, 80, 1)
		d.state["humidity"] = clamp(number(d.state["humidity"], 45)+(rand.Float64()*2-1)*humidity.change, humidity.min, humidity.max)
		d.state["co2"] = clamp(number(d.state["co2"], 600)+(rand.Float64()*2-1)*15, 400, 1200)
	case "binary_sensor":
		// Home Assistant style on/off text; state_probability is the
		// chance of the active state each tick, so 0 keeps it quiet.
		if key := firstStateKey(d.state); key != "" {
			if rand.Float64() < pointerFloat(d.simulation.StateProbability, 0) {
				d.state[key] = "on"
			} else {
				d.state[key] = "off"
			}
		}
	case "humidity_sensor":
		humidity := d.simulationRange(d.simulation.Humidity, 20, 80, 1)
		d.state["humidity"] = clamp(number(d.state["humidity"], 45)+(rand.Float64()*2-1)*humidity.change, humidity.min, humidity.max)
	case "motion_sensor":
		d.state["motion"] = rand.Float64() < pointerFloat(d.simulation.MotionProbability, 0.15)
	case "door_sensor":
		d.state["open"] = rand.Float64() < pointerFloat(d.simulation.OpenProbability, 0.05)
	case "air_conditioner", "smart_oven", "thermostat":
		if d.info.Type != "thermostat" && !boolean(d.state["on"]) {
			break
		}
		current := number(d.state["current_temperature"], 21)
		target := number(d.state["target_temperature"], 21)
		step := pointerFloat(d.simulation.TemperatureStep, 0.5)
		d.state["current_temperature"] = current + clamp(target-current, -step, step)
	case "smart_plug":
		if boolean(d.state["on"]) {
			power := d.simulationRange(d.simulation.Power, 60, 80, 0)
			d.state["power"] = power.min + rand.Float64()*(power.max-power.min)
		} else {
			d.state["power"] = 0.0
		}
	case "energy_sensor":
		powerSettings := d.simulationRange(d.simulation.Power, 50, 200, 5)
		power := number(d.state["power"], powerSettings.min)
		power += (rand.Float64()*2 - 1) * powerSettings.change
		d.state["power"] = clamp(power, powerSettings.min, powerSettings.max)
		d.state["energy"] = number(d.state["energy"], 0) + power*pointerFloat(d.simulation.EnergyPerTick, 1.0/360.0)
	}
	return d.readings(now)
}

type simulationRange struct {
	min, max, change float64
}

func (d *genericDevice) nextTemperature(current float64) float64 {
	settings := d.simulationRange(d.simulation.Temperature, 16, 28, 0.2)
	target := 21.0
	if settings.min > target {
		target = settings.min
	}
	if settings.max < target {
		target = settings.max
	}
	next := current + (target-current)*0.05 + (rand.Float64()*2-1)*settings.change
	return clamp(next, settings.min, settings.max)
}

func (d *genericDevice) simulationRange(settings config.NumericRange, defaultMin, defaultMax, defaultChange float64) simulationRange {
	min := pointerFloat(settings.Min, defaultMin)
	max := pointerFloat(settings.Max, defaultMax)
	if max < min {
		max = min
	}
	return simulationRange{min: min, max: max, change: pointerFloat(settings.MaxChange, defaultChange)}
}

func pointerFloat(value *float64, fallback float64) float64 {
	if value == nil {
		return fallback
	}
	return *value
}

// firstStateKey returns the lexically first state key, so single-property
// devices behave deterministically despite Go map ordering.
func firstStateKey(state map[string]any) string {
	keys := make([]string, 0, len(state))
	for key := range state {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return ""
	}
	return keys[0]
}

// textPowerState reports whether a device type reports its power state as
// Home Assistant style on/off text instead of a boolean.
func textPowerState(deviceType string) bool {
	return deviceType == "light" || deviceType == "fan" || deviceType == "siren"
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
		"smoke":               {"smoke", ""},
		"leak":                {"moisture", ""},
		"co2":                 {"carbon_dioxide", "ppm"},
		"locked":              {"lock", ""},
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
		entityID := d.info.ID + "." + property
		if id, ok := d.entities[property]; ok && id != "" {
			entityID = id
		}
		result = append(result, Reading{DeviceID: d.info.ID, Property: property, EntityID: entityID, State: state, Attributes: attributes, Time: now})
	}
	return result
}

func (d *genericDevice) ApplyCommand(command dto.Command) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.Addresses(command.ExternalEntityID) {
		return fmt.Errorf("command targets %q, device is %q", command.ExternalEntityID, d.info.ID)
	}
	switch command.Command {
	case "turn_on", "turn_off":
		if !hasCapability(d.info.Capabilities, command.Command) {
			return fmt.Errorf("device %q does not support %s", d.info.ID, command.Command)
		}
		if textPowerState(d.info.Type) {
			if command.Command == "turn_on" {
				d.state["on"] = "on"
			} else {
				d.state["on"] = "off"
			}
		} else {
			d.state["on"] = command.Command == "turn_on"
		}
	case "lock", "unlock":
		if !hasCapability(d.info.Capabilities, command.Command) {
			return fmt.Errorf("device %q does not support %s", d.info.ID, command.Command)
		}
		if command.Command == "lock" {
			d.state["locked"] = "locked"
		} else {
			d.state["locked"] = "unlocked"
		}
	case "open", "close":
		if !hasCapability(d.info.Capabilities, command.Command) {
			return fmt.Errorf("device %q does not support %s", d.info.ID, command.Command)
		}
		if command.Command == "open" {
			d.state["open"] = "open"
		} else {
			d.state["open"] = "closed"
		}
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

// addresses reports whether the command's external entity id targets this
// device: the device id itself, an entity id of one of its properties
// ("{device_id}.{property}", the form readings publish), or the legacy
// state topic.
func (d *genericDevice) Addresses(externalEntityID string) bool {
	if externalEntityID == d.info.ID || externalEntityID == d.topicID() {
		return true
	}
	for _, entityID := range d.entities {
		if externalEntityID == entityID {
			return true
		}
	}
	deviceID, _, found := strings.Cut(externalEntityID, ".")
	return found && deviceID == d.info.ID
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
