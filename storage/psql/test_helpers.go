package psql

import (
	"context"
	"fmt"

	"github.com/dannyvelas/parkspot-backend/config"
	"github.com/dannyvelas/parkspot-backend/migrations"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func NewSandboxDatabase() (testcontainers.Container, Database, error) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres",
		ExposedPorts: []string{"5432/tcp", "5432/tcp"},
		WaitingFor:   wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
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

	if err := migrations.Up(postgresURL); err != nil {
		_ = postgresContainer.Terminate(ctx)
		return nil, Database{}, fmt.Errorf("tearing down because failed to seed schemas: %v", err)
	}

	return postgresContainer, database, nil
}

// ResetTables deletes every row from the given tables, in the order given,
// so a test can start the next case from an empty table. Tables with
// foreign keys into one another must be listed child-before-parent.
func (database Database) ResetTables(tables ...string) error {
	for _, table := range tables {
		if _, err := database.driver.Exec("DELETE FROM " + table); err != nil {
			return fmt.Errorf("resetting table %s: %w", table, err)
		}
	}

	return nil
}
