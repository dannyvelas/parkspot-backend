package psql

import (
	"fmt"

	"github.com/dannyvelas/parkspot-backend/config"
	"github.com/dannyvelas/parkspot-backend/errs"
	"github.com/dannyvelas/parkspot-backend/storage"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
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
		return Database{}, fmt.Errorf("database: %w: %v", errs.ErrDBConnecting, err)
	}

	err = driver.Ping()
	if err != nil {
		return Database{}, fmt.Errorf("database: %w: %v", errs.ErrDBPinging, err)
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
