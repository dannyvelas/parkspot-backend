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
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		postgresConfig.Host(),
		postgresConfig.Port(),
		postgresConfig.User(),
		postgresConfig.Password(),
		postgresConfig.DbName(),
	)

	driver, err := sqlx.Open("postgres", psqlInfo)
	if err != nil {
		return Database{}, err
	}

	err = driver.Ping()
	if err != nil {
		return Database{}, err
	}

	return Database{driver: driver}, nil
}
