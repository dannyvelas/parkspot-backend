package config

type PostgresConfig struct {
	URL string
}

func newPostgresConfig() PostgresConfig {
	return PostgresConfig{
		URL: readEnvString("DATABASE_URL", "postgresql://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable"),
	}
}
