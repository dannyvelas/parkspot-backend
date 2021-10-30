package config

import "os"

type TokenConfig struct {
	secret string
}

func newTokenConfig() (*TokenConfig, error) {
	secret := os.Getenv("TOKEN_SECRET")
	if secret == "" {
		return nil, varNotFoundError("TOKEN_SECRET")
	}

	return &TokenConfig{
		secret: secret,
	}, nil
}

func (tokenConfig TokenConfig) Secret() string {
	return tokenConfig.secret
}
