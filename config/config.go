package config

import (
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"path/filepath"
	"runtime"
)

type Config struct {
	Http     HttpConfig
	Postgres PostgresConfig
	Token    TokenConfig
	OAuth    OAuthConfig
}

func NewConfig() (Config, error) {
	if err := loadDotEnv(); err != nil {
		log.Warn().Msgf("config: .env file not found: %v", err)
	}

	oauthConfig, err := newOAuthConfig()
	if err != nil {
		return Config{}, err
	}

	httpConfig, err := newHttpConfig()
	if err != nil {
		return Config{}, err
	}

	return Config{
		Http:     httpConfig,
		Postgres: newPostgresConfig(),
		Token:    newTokenConfig(),
		OAuth:    oauthConfig,
	}, nil
}

func loadDotEnv() error {
	rootPath := getRootPath()
	return godotenv.Load(filepath.Join(rootPath, ".env"))
}

func getRootPath() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..")
}
