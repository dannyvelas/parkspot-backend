package psql

import (
	"context"
	"fmt"
	"os"

	"github.com/dannyvelas/parkspot-backend/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/rs/zerolog/log"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// schemaOnlyMigrations lists migration files, beyond the base schema (version 1),
// that contain schema changes rather than seed data. CreateSchemas applies these
// directly (not via migrator.Migrate) because migrate.Migrate(N) walks every
// intermediate version sequentially, which would also apply the seed-data
// migrations (000002-000006) that tests deliberately skip.
var schemaOnlyMigrations = []string{
	"000007_quota_adjustment.up.sql",
}

func NewSandboxDatabase() (testcontainers.Container, Database, error) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres",
		ExposedPorts: []string{"5432/tcp", "5432/tcp"},
		WaitingFor:   wait.ForExposedPort(),
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "postgres",
			"POSTGRES_HOST":     "postgres",
		},
	}
	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, Database{}, fmt.Errorf("failed to set up container: %v", err)
	}

	// get the endpoint of the container we just created
	postgresEndpoint, err := postgresContainer.Endpoint(ctx, "")
	if err != nil {
		_ = postgresContainer.Terminate(ctx)
		return nil, Database{}, fmt.Errorf("tearing down because failed to get endpoint: %v", err)
	}
	postgresURL := fmt.Sprintf("postgresql://postgres:postgres@%s/postgres?sslmode=disable&connect_timeout=60", postgresEndpoint)

	database, err := NewDatabase(config.PostgresConfig{URL: postgresURL})
	if err != nil {
		_ = postgresContainer.Terminate(ctx)
		return nil, Database{}, fmt.Errorf("tearing down because failed to instantiate database: %v", err)
	}

	if err := CreateSchemasForTest(database); err != nil {
		_ = postgresContainer.Terminate(ctx)
		return nil, Database{}, fmt.Errorf("tearing down because failed to seed schemas: %v", err)
	}

	return postgresContainer, database, nil
}

func CreateSchemasForTest(database Database) error {
	driver, err := postgres.WithInstance(database.driver.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("call to postgres.WithInstance failed to cast *sql.DB to migrate.Driver: %v", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance("file://../migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to initialize migrate with migrate.Driver instance: %v", err)
	}

	if version, dirty, err := migrator.Version(); dirty {
		return fmt.Errorf("error: database version is dirty. Please fix it")
	} else if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("error getting migrator version: %v", err)
	} else if err != migrate.ErrNilVersion {
		log.Info().Msgf("not applying any migrations because found a version of %d", version)
	} else {
		if err := migrator.Migrate(1); err != nil {
			return fmt.Errorf("failed to migrate up to the first migration: %v", err)
		}

		for _, filename := range schemaOnlyMigrations {
			sqlBytes, err := os.ReadFile("../migrations/" + filename)
			if err != nil {
				return fmt.Errorf("failed to read schema-only migration %s: %v", filename, err)
			}
			if _, err := database.driver.Exec(string(sqlBytes)); err != nil {
				return fmt.Errorf("failed to apply schema-only migration %s: %v", filename, err)
			}
		}
	}

	return nil
}
