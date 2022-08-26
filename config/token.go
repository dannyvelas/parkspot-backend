package config

type TokenConfig struct {
	accessSecret  string
	refreshSecret string
}

func newTokenConfig() TokenConfig {
	return TokenConfig{
		accessSecret:  readEnvString("TOKEN_ACCESSSECRET", "accessSecret"),
		refreshSecret: readEnvString("TOKEN_REFRESHSECRET", "refreshSecret"),
	}
}

func (tokenConfig TokenConfig) AccessSecret() string {
	return tokenConfig.accessSecret
}

func (tokenConfig TokenConfig) RefreshSecret() string {
	return tokenConfig.refreshSecret
}
