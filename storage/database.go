package storage

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

type Database struct {
	driver *sqlx.DB
}

func NewPostgresDatabase(postgresConfig config.PostgresConfig) (Database, error) {
	driver, err := sqlx.Connect("postgres", postgresConfig.URL())
	if err != nil {
		return Database{}, fmt.Errorf("database: %w: %v", errs.DBConnecting, err)
	}

	return Database{driver: driver}, nil
}

func NewMemoryDatabase() (Database, error) {
	driver, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		return Database{}, fmt.Errorf("database: %w: %v", errs.DBConnecting, err)
	}

	err = seedMemoryDB(driver)
	if err != nil {
		return Database{}, fmt.Errorf("error seeding mock database: %v", err)
	}

	return Database{driver: driver}, nil
}

func seedMemoryDB(driver *sqlx.DB) error {
	migrateDriver, err := sqlite.WithInstance(driver.DB, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("Call to postgres.WithInstance failed to cast *sql.DB to migrate.Driver: %v", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance("file://./migrations", "mockdatabase", migrateDriver)
	if err != nil {
		return fmt.Errorf("Failed to initialize migrate with migrate.Driver instance: %v", err)
	}

	if version, dirty, err := migrator.Version(); dirty {
		return fmt.Errorf("Error: database version is dirty. Please fix it.")
	} else if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("Error getting migrator version: %v", err)
	} else if err == migrate.ErrNilVersion || version < 6 {
		if err := migrator.Up(); err != nil {
			return fmt.Errorf("Failed to migrate up: %v", err)
		}
	}

	return nil
}
