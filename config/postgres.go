package config

import (
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

func newPostgresConfig() (PostgresConfig, error) {
	var postgresConfig PostgresConfig

	postgresConfig.host = os.Getenv("PG_HOST")
	if postgresConfig.host == "" {
		return PostgresConfig{}, notFoundError{"PG_HOST"}
	}

	if portString := os.Getenv("PG_PORT"); portString == "" {
		return PostgresConfig{}, notFoundError{"PG_PORT"}
	} else if parsed, err := strconv.ParseUint(portString, 10, 64); err != nil {
		return PostgresConfig{}, conversionError{"PG_PORT", "uint"}
	} else if parsed < 80 {
		return PostgresConfig{}, invalidError{"PG_PORT", "< 80"}
	} else {
		postgresConfig.port = uint(parsed)
	}

	postgresConfig.user = os.Getenv("PG_USER")
	if postgresConfig.user == "" {
		return PostgresConfig{}, notFoundError{"PG_USER"}
	}

	postgresConfig.password = os.Getenv("PG_PASSWORD")
	if postgresConfig.password == "" {
		return PostgresConfig{}, notFoundError{"PG_PASSWORD"}
	}

	postgresConfig.dbName = os.Getenv("PG_DBNAME")
	if postgresConfig.dbName == "" {
		return PostgresConfig{}, notFoundError{"PG_DBNAME"}
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
