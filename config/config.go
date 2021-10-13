package config

import (
	"errors"
	"fmt"
)

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
}

func New() (*Config, error) {
	postgresConfig, err := NewPostgresConfig()
	if err != nil {
		return nil, err
	}

	serverConfig, err := NewServerConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		Server:   *serverConfig,
		Postgres: *postgresConfig,
	}, nil
}

func varNotFoundError(variable string) error {
	return errors.New(fmt.Sprintf("No config value found for %s", variable))
}
