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

func newPostgresConfig() (*PostgresConfig, error) {
	host := os.Getenv("PG_HOST")
	if host == "" {
		return nil, varNotFoundError("PG_HOST")
	}

	port := os.Getenv("PG_PORT")
	if port == "" {
		return nil, varNotFoundError("PG_PORT")
	}

	portAsInt, err := strconv.Atoi(port)
	if err != nil {
		return nil, errors.New("PG_PORT could not be converted to an integer.")
	}

	user := os.Getenv("PG_USER")
	if user == "" {
		return nil, varNotFoundError("PG_USER")
	}

	password := os.Getenv("PG_PASSWORD")
	if password == "" {
		return nil, varNotFoundError("PG_PASSWORD")
	}

	dbName := os.Getenv("PG_DBNAME")
	if dbName == "" {
		return nil, varNotFoundError("PG_DBNAME")
	}

	return &PostgresConfig{
		host:     host,
		port:     portAsInt,
		user:     user,
		password: password,
		dbName:   dbName,
	}, nil
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
