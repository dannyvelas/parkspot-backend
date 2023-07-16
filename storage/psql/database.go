package psql

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

type Database struct {
	driver       *sqlx.DB
	adminRepo    storage.AdminRepo
	residentRepo storage.ResidentRepo
	carRepo      storage.CarRepo
	permitRepo   storage.PermitRepo
	visitorRepo  storage.VisitorRepo
}

func NewDatabase(postgresConfig config.PostgresConfig) (Database, error) {
	driver, err := sqlx.Open("postgres", postgresConfig.URL)
	if err != nil {
		return Database{}, fmt.Errorf("database: %w: %v", errs.DBConnecting, err)
	}

	err = driver.Ping()
	if err != nil {
		return Database{}, fmt.Errorf("database: %w: %v", errs.DBPinging, err)
	}

	return Database{
		driver:       driver,
		adminRepo:    NewAdminRepo(driver),
		residentRepo: NewResidentRepo(driver),
		carRepo:      NewCarRepo(driver),
		permitRepo:   NewPermitRepo(driver),
		visitorRepo:  NewVisitorRepo(driver),
	}, nil
}

func (database Database) CreateSchemas() error {
	driver, err := postgres.WithInstance(database.driver.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("Call to postgres.WithInstance failed to cast *sql.DB to migrate.Driver: %v", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance("file://../migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("Failed to initialize migrate with migrate.Driver instance: %v", err)
	}

	if version, dirty, err := migrator.Version(); dirty {
		return fmt.Errorf("Error: database version is dirty. Please fix it.")
	} else if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("Error getting migrator version: %v", err)
	} else if err != migrate.ErrNilVersion {
		log.Info().Msgf("not applying any migrations because found a version of %d", version)
	} else {
		if err := migrator.Migrate(1); err != nil {
			return fmt.Errorf("Failed to migrate up to the first migration: %v", err)
		}
	}

	return nil
}

func (database Database) AdminRepo() storage.AdminRepo {
	return database.adminRepo
}
func (database Database) ResidentRepo() storage.ResidentRepo {
	return database.residentRepo
}
func (database Database) CarRepo() storage.CarRepo {
	return database.carRepo
}
func (database Database) PermitRepo() storage.PermitRepo {
	return database.permitRepo
}
func (database Database) VisitorRepo() storage.VisitorRepo {
	return database.visitorRepo
}
