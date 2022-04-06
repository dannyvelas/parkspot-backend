package storage

import (
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func GetMigrator(database Database) (*migrate.Migrate, error) {
	driver, err := postgres.WithInstance(database.driver, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("Failed to cast Database.driver to migrate.Driver interface: %v", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance("file://../migrations", "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize migrator: %v", err)
	}

	if err := migrator.Steps(1); err != nil {
		return nil, fmt.Errorf("Failed to go to v1 migrations: %v", err)
	}

	return migrator, nil
}
