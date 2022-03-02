package config

type TokenConfig struct {
	secret string
}

const (
	defaultTokenSecret = "THISISASECRET"
)

func newTokenConfig() TokenConfig {
	return TokenConfig{
		secret: readEnvString("TOKEN_SECRET", defaultTokenSecret),
	}
}

func (tokenConfig TokenConfig) Secret() string {
	return tokenConfig.secret
}
