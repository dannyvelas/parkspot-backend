package config

type TokenConfig struct {
	AccessSecret  string
	RefreshSecret string
}

func newTokenConfig() TokenConfig {
	return TokenConfig{
		AccessSecret:  readEnvString("TOKEN_ACCESSSECRET", "accessSecret"),
		RefreshSecret: readEnvString("TOKEN_REFRESHSECRET", "refreshSecret"),
	}
}
