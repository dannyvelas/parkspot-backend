package config

import (
	"os"
)

type TokenConfig struct {
	refresh string
	access  string
}

func newTokenConfig() (*TokenConfig, error) {
	refresh := os.Getenv("TOKEN_REFRESH")
	if refresh == "" {
		return nil, varNotFoundError("TOKEN_REFRESH")
	}

	access := os.Getenv("TOKEN_ACCESS")
	if access == "" {
		return nil, varNotFoundError("TOKEN_ACCESS")
	}

	return &TokenConfig{
		refresh: refresh,
		access:  access,
	}, nil
}

func (tokenConfig TokenConfig) Refresh() string {
	return tokenConfig.refresh
}

func (tokenConfig TokenConfig) Access() string {
	return tokenConfig.access
}
