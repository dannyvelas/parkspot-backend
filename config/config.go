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
	config.postgres = newPostgresConfig()
	config.token = newTokenConfig()

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
