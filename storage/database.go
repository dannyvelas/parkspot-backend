package storage

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Database struct {
	driver *sqlx.DB
}

func NewDatabase(postgresConfig config.PostgresConfig) (Database, error) {
	driver, err := sqlx.Connect("postgres", postgresConfig.URL())
	if err != nil {
		return Database{}, fmt.Errorf("database: %w: %v", ErrConnecting, err)
	}

	return Database{driver: driver}, nil
}
