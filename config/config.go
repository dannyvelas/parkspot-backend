package config

import (
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Config struct {
	token    TokenConfig
	postgres PostgresConfig
	http     HttpConfig
}

func New() (Config, error) {
	var config Config

	err := godotenv.Load()
	if err != nil {
		log.Warn().Msg(".env file not found.")
	}

	config.http = newHttpConfig()

	if postgresConfig, err := newPostgresConfig(); err != nil {
		return Config{}, err
	} else {
		config.postgres = postgresConfig
	}

	if tokenConfig, err := newTokenConfig(); err != nil {
		return Config{}, err
	} else {
		config.token = tokenConfig
	}

	return config, nil
}

func (config Config) Token() TokenConfig {
	return config.token
}

func (config Config) Postgres() PostgresConfig {
	return config.postgres
}

func (config Config) Http() HttpConfig {
	return config.http
}
