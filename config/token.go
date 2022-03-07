package config

type TokenConfig struct {
	secret string
}

const (
	defaultTokenSecret = "tokensecret"
)

func newTokenConfig() TokenConfig {
	return TokenConfig{
		secret: readEnvString("TOKEN_SECRET", defaultTokenSecret),
	}
}

func (tokenConfig TokenConfig) Secret() string {
	return tokenConfig.secret
}
