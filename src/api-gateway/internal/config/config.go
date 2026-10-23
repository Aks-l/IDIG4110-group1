package config

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"os"

	"IDIG4110/shared/env"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Server    ServerConfig     `yaml:"server"`
	Proxy     ProxyConfig      `yaml:"proxy"`
	Health    HealthConfig     `yaml:"health"`
	Upstreams []UpstreamConfig `yaml:"upstreams"`
	Auth      AuthConfig
}

type AuthConfig struct {
	PublicKey ed25519.PublicKey
}

func (ac *AuthConfig) Load() error {
	encoded, err := env.Required("AUTH_JWT_PUBLIC_KEY")
	if err != nil {
		return err
	}

	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("Deoced jwlt public key: %w", err)
	}

	if len(raw) != ed25519.PublicKeySize {
		return fmt.Errorf(
			"Invalid ED25519 public key length: got %d bytes, want %d",
			len(raw),
			ed25519.PublicKeySize,
		)
	}

	ac.PublicKey = ed25519.PublicKey(raw)

	return nil
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

	if err := cfg.Auth.Load(); err != nil {
		return nil, err
	}

	return &cfg, nil
}
