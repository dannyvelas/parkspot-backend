package config

import (
	"errors"
	"os"
	"strconv"
)

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Dbname   string
}

func NewPostgresConfig() (*PostgresConfig, error) {
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
		Host:     host,
		Port:     portAsInt,
		User:     user,
		Password: password,
		Dbname:   dbName,
	}, nil
}
