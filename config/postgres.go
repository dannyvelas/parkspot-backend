package config

import (
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
)

type PostgresConfig struct {
	host     string
	port     uint
	user     string
	password string
	dbName   string
}

const (
	defaultHost         = "postgres"
	defaultPostgresPort = 5432
	defaultUser         = "postgres"
	defaultPassword     = "postgres"
	defaultDbName       = "postgres"
)

func newPostgresConfig() (PostgresConfig, error) {
	var postgresConfig PostgresConfig

	postgresConfig.host = os.Getenv("PG_HOST")
	if postgresConfig.host == "" {
		log.Warn().Msg(notFoundError{"PG_HOST"}.ErrorUsingDefault(defaultHost))
		postgresConfig.host = defaultHost
	}

	if portString := os.Getenv("PG_PORT"); portString == "" {
		log.Warn().Msg(notFoundError{"PG_PORT"}.ErrorUsingDefault(defaultPostgresPort))
		postgresConfig.port = defaultPostgresPort
	} else if parsed, err := strconv.ParseUint(portString, 10, 64); err != nil {
		log.Warn().Msg(conversionError{"PG_PORT", "uint"}.ErrorUsingDefault(defaultPostgresPort))
		postgresConfig.port = defaultPostgresPort
	} else {
		postgresConfig.port = uint(parsed)
	}

	postgresConfig.user = os.Getenv("PG_USER")
	if postgresConfig.user == "" {
		log.Warn().Msg(notFoundError{"PG_USER"}.ErrorUsingDefault(defaultUser))
		postgresConfig.user = defaultUser
	}

	postgresConfig.password = os.Getenv("PG_PASSWORD")
	if postgresConfig.password == "" {
		log.Warn().Msg(notFoundError{"PG_PASSWORD"}.ErrorUsingDefault(defaultPassword))
		postgresConfig.password = defaultPassword
	}

	postgresConfig.dbName = os.Getenv("PG_DBNAME")
	if postgresConfig.dbName == "" {
		log.Warn().Msg(notFoundError{"PG_DBNAME"}.ErrorUsingDefault(defaultDbName))
		postgresConfig.dbName = defaultDbName
	}

	return postgresConfig, nil
}

func (postgresConfig PostgresConfig) Host() string {
	return postgresConfig.host
}

func (postgresConfig PostgresConfig) Port() uint {
	return postgresConfig.port
}

func (postgresConfig PostgresConfig) User() string {
	return postgresConfig.user
}

func (postgresConfig PostgresConfig) Password() string {
	return postgresConfig.password
}

func (postgresConfig PostgresConfig) DbName() string {
	return postgresConfig.dbName
}
