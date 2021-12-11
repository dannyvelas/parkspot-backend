package storage

import (
	"database/sql"
	"fmt"
	"github.com/dannyvelas/parkspot-api/config"
	_ "github.com/lib/pq"
)

type Database struct {
	driver *sql.DB
}

func NewDatabase(postgres_config config.PostgresConfig) (*Database, error) {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		postgres_config.Host(),
		postgres_config.Port(),
		postgres_config.User(),
		postgres_config.Password(),
		postgres_config.DbName(),
	)

	driver, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = driver.Ping()
	if err != nil {
		return nil, err
	}

	return &Database{driver: driver}, nil
}
