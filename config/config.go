package config

import (
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Config struct {
	HTTP     HTTPConfig
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

	httpConfig, err := newHTTPConfig()
	if err != nil {
		return Config{}, err
	}

	return Config{
		HTTP:     httpConfig,
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
