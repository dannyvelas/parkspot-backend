package config

import (
	"errors"
	"fmt"
)

type Config struct {
	server   ServerConfig
	postgres PostgresConfig
}

func New() (*Config, error) {
	postgresConfig, err := newPostgresConfig()
	if err != nil {
		return nil, err
	}

	serverConfig, err := newServerConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		server:   *serverConfig,
		postgres: *postgresConfig,
	}, nil
}

func (config Config) Server() ServerConfig {
	return config.server
}

func (config Config) Postgres() PostgresConfig {
	return config.postgres
}

func varNotFoundError(variable string) error {
	return errors.New(fmt.Sprintf("No config value found for %s", variable))
}
