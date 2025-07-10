package psql

import (
	"context"
	"fmt"
	"github.com/dannyvelas/parkspot-backend/config"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

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
		return nil, Database{}, fmt.Errorf("Failed to set up container: %v", err)
	}

	// get the endpoint of the container we just created
	postgresEndpoint, err := postgresContainer.Endpoint(ctx, "")
	if err != nil {
		postgresContainer.Terminate(ctx)
		return nil, Database{}, fmt.Errorf("tearing down because failed to get endpoint: %v", err)
	}
	postgresURL := fmt.Sprintf("postgresql://postgres:postgres@%s/postgres?sslmode=disable&connect_timeout=60", postgresEndpoint)

	database, err := NewDatabase(config.PostgresConfig{URL: postgresURL})
	if err != nil {
		postgresContainer.Terminate(ctx)
		return nil, Database{}, fmt.Errorf("tearing down because failed to instantiate database: %v", err)
	}

	if err := database.CreateSchemas(); err != nil {
		postgresContainer.Terminate(ctx)
		return nil, Database{}, fmt.Errorf("tearing down because failed to seed schemas: %v", err)
	}

	return postgresContainer, database, nil
}
