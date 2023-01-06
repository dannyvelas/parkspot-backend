package config

import (
	"flag"
	"github.com/dannyvelas/lasvistas_api/util"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"os"
	"regexp"
)

type Config struct {
	http              HttpConfig
	postgres          PostgresConfig
	token             TokenConfig
	oauth             OAuthConfig
	UseMemoryDatabase bool
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
	// determine whether to use in-memory database
	useMemoryDatabase := flag.Bool("memory", false, "if present, an in-memory database will be used")
	if useMemoryDatabase == nil {
		useMemoryDatabase = util.ToPtr(false)
	}
	flag.Parse()

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
		http:              httpConfig,
		postgres:          newPostgresConfig(),
		token:             newTokenConfig(),
		oauth:             oauthConfig,
		UseMemoryDatabase: *useMemoryDatabase,
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
