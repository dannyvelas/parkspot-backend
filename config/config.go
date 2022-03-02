package config

import (
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Config struct {
	http     HttpConfig
	postgres PostgresConfig
	token    TokenConfig
}

func NewConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Warn().Msg(".env file not found.")
	}

	return Config{
		http:     newHttpConfig(),
		postgres: newPostgresConfig(),
		token:    newTokenConfig(),
	}
}

func (config Config) Http() HttpConfig {
	return config.http
}

func (config Config) Postgres() PostgresConfig {
	return config.postgres
}

func (config Config) Token() TokenConfig {
	return config.token
}
