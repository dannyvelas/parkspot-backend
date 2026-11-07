package migrations

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func Up(databaseURL string) error {
	source, err := iofs.New(FS, ".")
	if err != nil {
		return fmt.Errorf("migrations.Up: failed to create source instance: %v", err)
	}

	migrator, err := migrate.NewWithSourceInstance("iofs", source, databaseURL)
	if err != nil {
		return fmt.Errorf("migrations.Up: failed to initialize migrator: %v", err)
	}

	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrations.Up: failed to apply migrations: %v", err)
	}

	return nil
}
