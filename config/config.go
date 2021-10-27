package config

import (
	"errors"
	"fmt"
)

type Config struct {
	token    TokenConfig
	postgres PostgresConfig
}

func New() (*Config, error) {
	postgresConfig, err := newPostgresConfig()
	if err != nil {
		return nil, err
	}

	tokenConfig, err := newTokenConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		token:    *tokenConfig,
		postgres: *postgresConfig,
	}, nil
}

func (config Config) Token() TokenConfig {
	return config.token
}

func (config Config) Postgres() PostgresConfig {
	return config.postgres
}

func varNotFoundError(variable string) error {
	return errors.New(fmt.Sprintf("No config value found for %s", variable))
}
