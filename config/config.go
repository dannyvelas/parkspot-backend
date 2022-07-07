package config

import (
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"os"
	"regexp"
)

type Config struct {
	http      HttpConfig
	postgres  PostgresConfig
	token     TokenConfig
	oauth     OAuthConfig
	constants Constants
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

	return Config{
		http:      newHttpConfig(),
		postgres:  newPostgresConfig(),
		token:     newTokenConfig(),
		oauth:     oauthConfig,
		constants: newConstants(),
	}, nil
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

func (config Config) OAuth() OAuthConfig {
	return config.oauth
}

func (config Config) Constants() Constants {
	return config.constants
}
