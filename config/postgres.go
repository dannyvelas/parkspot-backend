package config

type PostgresConfig struct {
	url string
}

func newPostgresConfig() PostgresConfig {
	return PostgresConfig{
		url: readEnvString("DATABASE_URL", "postgres://postgres:postgres@127.0.0.1:5432/postgres"),
	}
}

func (postgresConfig PostgresConfig) URL() string {
	return postgresConfig.url
}
