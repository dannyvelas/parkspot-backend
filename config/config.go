package config

type Config struct {
	token    TokenConfig
	postgres PostgresConfig
	http     HttpConfig
}

func New() (Config, error) {
	var config Config

	if httpConfig, err := newHttpConfig(); err != nil {
		return Config{}, err
	} else {
		config.http = httpConfig
	}

	if postgresConfig, err := newPostgresConfig(); err != nil {
		return Config{}, err
	} else {
		config.postgres = postgresConfig
	}

	if tokenConfig, err := newTokenConfig(); err != nil {
		return Config{}, err
	} else {
		config.token = tokenConfig
	}

	return config, nil
}

func (config Config) Token() TokenConfig {
	return config.token
}

func (config Config) Postgres() PostgresConfig {
	return config.postgres
}

func (config Config) Http() HttpConfig {
	return config.http
}
