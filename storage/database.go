package storage

import (
	"database/sql"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	_ "github.com/lib/pq"
)

type Database struct {
	driver *sql.DB
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

	driver, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return Database{}, err
	}

	err = driver.Ping()
	if err != nil {
		return Database{}, err
	}

	return Database{driver: driver}, nil
}
