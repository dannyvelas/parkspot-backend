package config

import (
	"os"
)

type ServerConfig struct {
	Address string
}

func newServerConfig() (*ServerConfig, error) {
	address := os.Getenv("SERVER_ADDRESS")
	if address == "" {
		return nil, varNotFoundError("SERVER_ADDRESS")
	}

	return &ServerConfig{
		Address: address,
	}, nil
}
