package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Server    ServerConfig     `yaml:"server"`
	Proxy     ProxyConfig      `yaml:"proxy"`
	Health    HealthConfig     `yaml:"health"`
	Upstreams []UpstreamConfig `yaml:"upstreams"`
}

type UpstreamConfig struct {
	Name   string `yaml:"name"`
	Prefix string `yaml:"prefix"`
	Url    string `yaml:"url"`
}

type ProxyConfig struct {
	ResponseHeaderTimeout int `yaml:"responseHeaderTimeout"`
}

type HealthConfig struct {
	Timeout int `yaml:"timeout"`
}

type ServerConfig struct {
	Port            int `yaml:"port"`
	ReadTimeout     int `yaml:"readTimeout"`
	WriteTimout     int `yaml:"writeTimout"`
	IdleTimeout     int `yaml:"idleTimout"`
	MaxHeaderBytes  int `yaml:"maxHeaderBytes"`
	ShutdownTimeout int `yaml:"shutdownTimeout"`
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

	if cfg.Proxy.ResponseHeaderTimeout <= 0 {
		cfg.Proxy.ResponseHeaderTimeout = 30
	}
	if cfg.Health.Timeout <= 0 {
		cfg.Health.Timeout = 3
	}
	if cfg.Server.ShutdownTimeout <= 0 {
		cfg.Server.ShutdownTimeout = 10
	}

	return &cfg, nil
}
