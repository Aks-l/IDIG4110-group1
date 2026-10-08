package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	Mqtt      MqttConfig      `yaml:"mqtt"`
	Migration MigrationConfig `yaml:"migration"`
	Kafka     KafkaConfig     `yaml:"kafka"`
}

// KafkaConfig connects ingest to the event bus. With no brokers, readings are
// only stored and commands are not forwarded.
type KafkaConfig struct {
	Brokers       string `yaml:"brokers"`
	CommandsGroup string `yaml:"commandsGroup"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
	SSL      string `yaml:"ssl"`
}

type ServerConfig struct {
	Port           int `yaml:"port"`
	ReadTimeout    int `yaml:"readTimeout"`
	WriteTimout    int `yaml:"writeTimout"`
	IdleTimeout    int `yaml:"idleTimout"`
	MaxHeaderBytes int `yaml:"maxHeaderBytes"`
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
	WorkerBufferSize     int    `yaml:"workerBufferSize"`
	Topic                string `yaml:"topic"`
}

type MigrationConfig struct {
	Directory string `yaml:"directory"`
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

	return &cfg, nil
}
