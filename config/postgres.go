package config

type PostgresConfig struct {
	url string
}

func newPostgresConfig() PostgresConfig {
	return PostgresConfig{
		url: readEnvString("DATABASE_URL", "postgresql://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable"),
	}
}

func (postgresConfig PostgresConfig) URL() string {
	return postgresConfig.url
}
