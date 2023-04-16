package config

import (
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"os"
	"regexp"
)

type Config struct {
	Http     HttpConfig
	Postgres PostgresConfig
	Token    TokenConfig
	OAuth    OAuthConfig
}

const projectName = "go-lasvistas_api"

func loadDotEnv() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	re := regexp.MustCompile(`^(.*` + projectName + `)`)
	rootPath := re.Find([]byte(cwd))

	return godotenv.Load(string(rootPath) + `/.env`)
}

func NewConfig() (Config, error) {
	err := loadDotEnv()
	if err != nil {
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
