package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Mqtt MqttConfig `yaml:"mqtt"`
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
	WorkerCount          int    `yaml:"workerCount"`
	WorkerBufferSize     int    `yaml:"WorkerBufferSize"`
	Topic                string `yaml:"topic"`
}

func Load(configPath string) (*Config, error) {
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(file,&cfg); err != nil {
		return nil, fmt.Errorf("paring yaml file: %w", err)
	}

	return &cfg, nil
}
