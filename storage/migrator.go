package storage

import (
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const highestVersion = 5

func GetUpMigrator(database Database) (*migrate.Migrate, error) {
	driver, err := postgres.WithInstance(database.driver.DB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("Failed to cast Database.driver to migrate.Driver interface: %v", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance("file://../migrations", "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize migrator: %v", err)
	}

	if version, dirty, err := migrator.Version(); dirty {
		return nil, fmt.Errorf("Error: database version is dirty. Please fix it.")
	} else if err != nil && err != migrate.ErrNilVersion {
		return nil, fmt.Errorf("Error getting migrator version: %v", err)
	} else if err == migrate.ErrNilVersion || version < highestVersion {
		if err := migrator.Up(); err != nil {
			return nil, fmt.Errorf("Failed to migrate up: %v", err)
		}
	}

	return migrator, nil
}
