package config

import (
	"errors"
	"os"
	"strconv"
)

type PostgresConfig struct {
	host     string
	port     int
	user     string
	password string
	dbName   string
}

func newPostgresConfig() (PostgresConfig, error) {
	var postgresConfig PostgresConfig

	postgresConfig.host = os.Getenv("PG_HOST")
	if postgresConfig.host == "" {
		return PostgresConfig{}, varNotFoundError("PG_HOST")
	}

	if portString := os.Getenv("PG_PORT"); portString == "" {
		return PostgresConfig{}, varNotFoundError("PG_PORT")
	} else if portInt, err := strconv.Atoi(portString); err != nil {
		return PostgresConfig{}, errors.New("PG_PORT could not be converted to an integer.")
	} else {
		postgresConfig.port = portInt
	}

	postgresConfig.user = os.Getenv("PG_USER")
	if postgresConfig.user == "" {
		return PostgresConfig{}, varNotFoundError("PG_USER")
	}

	postgresConfig.password = os.Getenv("PG_PASSWORD")
	if postgresConfig.password == "" {
		return PostgresConfig{}, varNotFoundError("PG_PASSWORD")
	}

	postgresConfig.dbName = os.Getenv("PG_DBNAME")
	if postgresConfig.dbName == "" {
		return PostgresConfig{}, varNotFoundError("PG_DBNAME")
	}

	return postgresConfig, nil
}

func (postgresConfig PostgresConfig) Host() string {
	return postgresConfig.host
}

func (postgresConfig PostgresConfig) Port() int {
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
