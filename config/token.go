package config

import "os"

type TokenConfig struct {
	secret string
}

func newTokenConfig() (TokenConfig, error) {
	var tokenConfig TokenConfig

	tokenConfig.secret = os.Getenv("TOKEN_SECRET")
	if tokenConfig.secret == "" {
		return TokenConfig{}, NotFoundError{"TOKEN_SECRET"}
	}

	return tokenConfig, nil
}

func (tokenConfig TokenConfig) Secret() string {
	return tokenConfig.secret
}
