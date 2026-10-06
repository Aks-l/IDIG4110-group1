package config

import (
	"IDIG4110/shared/env"
	"fmt"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Migration Migration
}

type ServerConfig struct {
	Port           int
	ReadTimeout    int
	WriteTimeout   int
	IdleTimeout    int
	MaxHeaderBytes int
}

type DatabaseConfig struct {
	URL string
}

type Migration struct {
	Directory string
}

func (db *DatabaseConfig) init() {
	user := env.Get("AUTH_DB_USER", "auth")
	port := env.GetInt("AUTH_DB_PORT", 5432)
	host := env.Get("AUTH_DB_HOST", "localhost")
	pwd := env.Get("AUTH_DB_PASSWORD", "123")
	ssl := env.Get("AUTH_DB_SSL", "disable")
	name := env.Get("AUTH_DB_NAME", "auth_db")

	db.URL = fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		user,
		pwd,
		host,
		port,
		name,
		ssl,
	)

	fmt.Printf(db.URL)
}

func (s *ServerConfig) init() {
	s.Port = env.GetInt("SERVER_PORT", 8080)
	s.ReadTimeout = env.GetInt("SERVER_READ_TIMEOUT", 15)
	s.WriteTimeout = env.GetInt("SERVER_WRITE_TIMEOUT", 15)
	s.IdleTimeout = env.GetInt("SERVER_IDLE_TIMEOUT", 60)
	s.MaxHeaderBytes = env.GetInt("SERVER_MAX_HEADER_BYTES", 1048576)
}

func (m *Migration) init() {
	m.Directory = env.Get("MIGRATION", "migrations")
}

func (c *Config) Init() {
	c.Server.init()
	c.Database.init()
	c.Migration.init()
}
