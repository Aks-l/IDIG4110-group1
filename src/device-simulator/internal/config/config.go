package config

import (
	"encoding/json"
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Mqtt        MqttConfig `yaml:"mqtt"`
	HouseConfig string     `yaml:"houseConfig"`
}

type MqttConfig struct {
	Url                  string `yaml:"url"`
	ClientId             string `yaml:"clientId"`
	AutoConnect          bool   `yaml:"autoConnect"`
	ConnectRetry         bool   `yaml:"connectRetry"`
	ConnectRetryInterval int    `yaml:"connectRetryInterval"`
	KeepAlive            int    `yaml:"keepAlive"`
	PingTimeout          int    `yaml:"pingTimeout"`
	CleanSession         bool   `yaml:"cleanSession"`
	Topic                string `yaml:"topic"`
	Interval             int    `yaml:"interval"`
	GatewayID            string `yaml:"gatewayId"`
}

type HouseConfig struct {
	House House `json:"house"`
}

type House struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	GatewayID string `json:"gateway_id"`
	Rooms     []Room `json:"rooms"`
}

type Room struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Devices []Device `json:"devices"`
}

type Device struct {
	ID           string           `json:"id"`
	Type         string           `json:"type"`
	InitialState map[string]any   `json:"initial_state"`
	Simulation   SimulationConfig `json:"simulation"`
}

type SimulationConfig struct {
	Enabled           *bool        `json:"enabled"`
	Temperature       NumericRange `json:"temperature"`
	Humidity          NumericRange `json:"humidity"`
	MotionProbability *float64     `json:"motion_probability"`
	OpenProbability   *float64     `json:"open_probability"`
	Power             NumericRange `json:"power"`
	EnergyPerTick     *float64     `json:"energy_per_tick"`
	TemperatureStep   *float64     `json:"temperature_step"`
}

type NumericRange struct {
	Min       *float64 `json:"min"`
	Max       *float64 `json:"max"`
	MaxChange *float64 `json:"max_change"`
}

func Load(configPath string) (*Config, error) {
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return nil, fmt.Errorf("parsing yaml file: %w", err)
	}
	if cfg.Mqtt.Topic == "" {
		return nil, fmt.Errorf("mqtt topic must not be empty")
	}
	if cfg.Mqtt.GatewayID == "" {
		return nil, fmt.Errorf("mqtt gatewayId must not be empty")
	}
	if cfg.HouseConfig == "" {
		return nil, fmt.Errorf("houseConfig must not be empty")
	}
	return &cfg, nil
}

func LoadHouse(configPath string) (*HouseConfig, error) {
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading house config: %w", err)
	}

	var cfg HouseConfig
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, fmt.Errorf("parsing house config JSON: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c HouseConfig) Validate() error {
	if c.House.ID == "" || c.House.Name == "" || c.House.GatewayID == "" {
		return fmt.Errorf("house id, name and gateway_id are required")
	}
	seen := map[string]bool{}
	for _, room := range c.House.Rooms {
		if room.ID == "" || room.Name == "" {
			return fmt.Errorf("room id and name are required")
		}
		for _, device := range room.Devices {
			if device.ID == "" || device.Type == "" {
				return fmt.Errorf("device id and type are required in room %q", room.ID)
			}
			if seen[device.ID] {
				return fmt.Errorf("duplicate device id %q", device.ID)
			}
			if err := device.Simulation.Validate(device.ID); err != nil {
				return err
			}
			seen[device.ID] = true
		}
	}
	return nil
}

func (s SimulationConfig) Validate(deviceID string) error {
	if err := s.Temperature.Validate(deviceID + ".simulation.temperature"); err != nil {
		return err
	}
	if err := s.Humidity.Validate(deviceID + ".simulation.humidity"); err != nil {
		return err
	}
	if err := s.Power.Validate(deviceID + ".simulation.power"); err != nil {
		return err
	}
	for name, value := range map[string]*float64{
		"motion_probability": s.MotionProbability,
		"open_probability":   s.OpenProbability,
	} {
		if value != nil && (*value < 0 || *value > 1) {
			return fmt.Errorf("%s.%s must be between 0 and 1", deviceID, name)
		}
	}
	for name, value := range map[string]*float64{
		"energy_per_tick":  s.EnergyPerTick,
		"temperature_step": s.TemperatureStep,
	} {
		if value != nil && *value < 0 {
			return fmt.Errorf("%s.%s must not be negative", deviceID, name)
		}
	}
	return nil
}

func (r NumericRange) Validate(name string) error {
	if r.Min != nil && r.Max != nil && *r.Min > *r.Max {
		return fmt.Errorf("%s.min must not be greater than max", name)
	}
	if r.MaxChange != nil && *r.MaxChange < 0 {
		return fmt.Errorf("%s.max_change must not be negative", name)
	}
	return nil
}
