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

func NewDatabase(postgresConfig config.PostgresConfig) (Database, error) {
	driver, err := sqlx.Connect("postgres", postgresConfig.URL())
	if err != nil {
		return Database{}, fmt.Errorf("database: %w: %v", errs.DBConnecting, err)
	}

	return Database{driver: driver}, nil
}

func NewMockDatabase() (Database, error) {
	driver, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		return Database{}, fmt.Errorf("database: %w: %v", errs.DBConnecting, err)
	}

	err = seedMockDB(driver)
	if err != nil {
		return Database{}, fmt.Errorf("error seeding mock database: %v", err)
	}

	return Database{driver: driver}, nil
}

func seedMockDB(driver *sqlx.DB) error {
	migrateDriver, err := sqlite.WithInstance(driver.DB, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("Call to postgres.WithInstance failed to cast *sql.DB to migrate.Driver: %v", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance("file://./migrations", "postgres", migrateDriver)
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

//func seedMockDB(driver *sqlx.DB) error {
//	type PreparedStmt struct {
//		stmt string
//		args []interface{}
//	}
//	stmts := []PreparedStmt{
//		{stmt: "CREATE TABLE IF NOT EXISTS admin(id TEXT PRIMARY KEY UNIQUE NOT NULL, first_name TEXT NOT NULL, last_name TEXT NOT NULL, email VARCHAR(255) UNIQUE NOT NULL, password VARCHAR(255) NOT NULL, is_privileged BOOLEAN NOT NULL, token_version INTEGER NOT NULL DEFAULT 0);"},
//		{stmt: "CREATE TABLE IF NOT EXISTS resident(id CHAR(8) PRIMARY KEY UNIQUE NOT NULL, first_name TEXT NOT NULL, last_name TEXT NOT NULL, phone VARCHAR(20) NOT NULL, email VARCHAR(255) NOT NULL, password VARCHAR(255) NOT NULL, unlim_days BOOLEAN NOT NULL DEFAULT FALSE, amt_parking_days_used SMALLINT NOT NULL DEFAULT 0, token_version INTEGER NOT NULL DEFAULT 0);"},
//		{stmt: "CREATE TABLE IF NOT EXISTS car(id UUID PRIMARY KEY UNIQUE NOT NULL, resident_id CHAR(8) NOT NULL, license_plate VARCHAR(10) UNIQUE NOT NULL, color TEXT NOT NULL, make TEXT, model TEXT, amt_parking_days_used SMALLINT NOT NULL DEFAULT 0);"},
//		{stmt: "CREATE TABLE IF NOT EXISTS permit(id SERIAL PRIMARY KEY UNIQUE NOT NULL, resident_id CHAR(8) NOT NULL, car_id UUID NOT NULL, license_plate VARCHAR(10) NOT NULL, color TEXT NOT NULL, make TEXT, model TEXT, start_ts BIGINT NOT NULL, end_ts BIGINT NOT NULL, request_ts BIGINT, affects_days BOOLEAN NOT NULL, exception_reason TEXT);"},
//		{`INSERT INTO admin( id, first_name, last_name, email, password, is_privileged) VALUES( $1, $2, $3, $4, $5, $6);`,
//			[]interface{}{
//				"test",
//				"Daniel",
//				"Velasquez",
//				"email@example.com",
//				"$2a$12$RwfoAooW.NM6Gj6j6BeqC.NpXCfOmdmIzGf3BrmMwfm7bdS5q7yty",
//				true,
//			},
//		}}
//	for _, tuple := range stmts {
//		_, err := driver.Exec(tuple.stmt, tuple.args...)
//		if err != nil {
//			return fmt.Errorf("sql.Exec: Error: %s\n", err)
//		}
//	}
//	return nil
//}
