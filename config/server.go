package config

import (
	"os"
)

type ServerConfig struct {
	address string
}

func newServerConfig() (*ServerConfig, error) {
	address := os.Getenv("SERVER_ADDRESS")
	if address == "" {
		return nil, varNotFoundError("SERVER_ADDRESS")
	}

	return &ServerConfig{
		address: address,
	}, nil
}

func (serverConfig ServerConfig) Address() string {
	return serverConfig.address
}
