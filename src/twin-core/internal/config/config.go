package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	Migration MigrationConfig `yaml:"migration"`
	Kafka     KafkaConfig     `yaml:"kafka"`
}

// KafkaConfig connects twin-core to the event bus. With no brokers,
// twin_state only changes through POST /api/v1/readings.
type KafkaConfig struct {
	Brokers string `yaml:"brokers"`
	Group   string `yaml:"group"`
}

type ServerConfig struct {
	Port           int `yaml:"port"`
	ReadTimeout    int `yaml:"readTimeout"`
	WriteTimeout   int `yaml:"writeTimeout"`
	IdleTimeout    int `yaml:"idleTimeout"`
	MaxHeaderBytes int `yaml:"maxHeaderBytes"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
	SSL      string `yaml:"ssl"`
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
