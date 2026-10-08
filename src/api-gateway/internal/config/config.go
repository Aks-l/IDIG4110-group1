package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Upstreams []UpstreamConfig `yaml:"upstreams"`
}

type UpstreamConfig struct {
	Name   string `yaml:"name"`
	Prefix string `yaml:"prefix"`
	Url    string `yaml:"url"`
}


type ServerConfig struct {
	Port           int `yaml:"port"`
	ReadTimeout    int `yaml:"readTimeout"`
	WriteTimout    int `yaml:"writeTimout"`
	IdleTimeout    int `yaml:"idleTimout"`
	MaxHeaderBytes int `yaml:"maxHeaderBytes"`
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
