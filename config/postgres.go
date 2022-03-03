package config

type PostgresConfig struct {
	host     string
	port     uint
	user     string
	password string
	dbName   string
}

const (
	defaultPostgresHost     = "127.0.0.1"
	defaultPostgresPort     = 5432
	defaultPostgresUser     = "postgres"
	defaultPostgresPassword = "postgres"
	defaultPostgresDbName   = "postgres"
)

func newPostgresConfig() PostgresConfig {
	return PostgresConfig{
		host:     readEnvString("PG_HOST", defaultPostgresHost),
		port:     readEnvUint("PG_PORT", defaultPostgresPort),
		user:     readEnvString("PG_USER", defaultPostgresUser),
		password: readEnvString("PG_PASSWORD", defaultPostgresPassword),
		dbName:   readEnvString("PG_DBNAME", defaultPostgresDbName),
	}
}

func (postgresConfig PostgresConfig) Host() string {
	return postgresConfig.host
}

func (postgresConfig PostgresConfig) Port() uint {
	return postgresConfig.port
}

func (postgresConfig PostgresConfig) User() string {
	return postgresConfig.user
}

func (postgresConfig PostgresConfig) Password() string {
	return postgresConfig.password
}

func (postgresConfig PostgresConfig) DbName() string {
	return postgresConfig.dbName
}
