// Package config reads the rules-engine settings from environment variables,
// so the same image runs in Docker Compose and in Kubernetes.
package config

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"IDIG4110/shared/env"
	"IDIG4110/shared/kafka"
)

type Config struct {
	HTTPPort        int
	DatabaseURL     string
	MigrationsDir   string
	KafkaBrokers    []string
	KafkaGroup      string
	CommandTTL      time.Duration
	TickInterval    time.Duration
	RefreshInterval time.Duration
}

// Load reads the configuration. DB_URL, if set, wins over the separate
// DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME and DB_SSLMODE variables.
func Load() (Config, error) {
	port, err := intVar("HTTP_PORT", 8080)
	if err != nil {
		return Config{}, err
	}
	ttl, err := intVar("COMMAND_TTL_SECONDS", 30)
	if err != nil {
		return Config{}, err
	}
	tick, err := intVar("TICK_MILLIS", 1000)
	if err != nil {
		return Config{}, err
	}
	refresh, err := intVar("RULE_REFRESH_SECONDS", 30)
	if err != nil {
		return Config{}, err
	}

	dbURL := env.Get("DB_URL", "")
	if dbURL == "" {
		u := url.URL{
			Scheme:   "postgres",
			User:     url.UserPassword(env.Get("DB_USER", "rules"), env.Get("DB_PASSWORD", "rules")),
			Host:     env.Get("DB_HOST", "rules-db") + ":" + env.Get("DB_PORT", "5432"),
			Path:     env.Get("DB_NAME", "rules_db"),
			RawQuery: "sslmode=" + env.Get("DB_SSLMODE", "disable"),
		}
		dbURL = u.String()
	}

	brokers := kafka.ParseBrokers(env.Get("KAFKA_BROKERS", "kafka:29092"))
	if len(brokers) == 0 {
		return Config{}, fmt.Errorf("KAFKA_BROKERS is empty")
	}

	return Config{
		HTTPPort:        port,
		DatabaseURL:     dbURL,
		MigrationsDir:   env.Get("MIGRATIONS_DIR", "migrations"),
		KafkaBrokers:    brokers,
		KafkaGroup:      env.Get("KAFKA_GROUP", "rules-engine"),
		CommandTTL:      time.Duration(ttl) * time.Second,
		TickInterval:    time.Duration(tick) * time.Millisecond,
		RefreshInterval: time.Duration(refresh) * time.Second,
	}, nil
}

func intVar(key string, fallback int) (int, error) {
	v := env.Get(key, "")
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer, got %q", key, v)
	}
	return n, nil
}
