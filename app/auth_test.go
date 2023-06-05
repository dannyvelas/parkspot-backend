package app

import (
	"context"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage/psql"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"testing"
)

type authTestSuite struct {
	suite.Suite
	container       testcontainers.Container
	authService     AuthService
	residentService ResidentService // kept here so we can tear down between tests
	resident        models.Resident // will exist in db for duration of tests
}

func TestAuthService(t *testing.T) {
	suite.Run(t, new(authTestSuite))
}

func (suite *authTestSuite) SetupSuite() {
	// configure and start container
	container, database, err := getSandboxDatabase()
	if err != nil {
		suite.T().Fatalf("error getting sandbox database: %v", err)
	}
	// save container in suite struct so we can terminate it on suite teardown
	suite.container = container

	// create services that are needed to create services used in this suite
	jwtService := NewJWTService(config.TokenConfig{AccessSecret: "accessSecret", RefreshSecret: "refereshSecret"})
	adminService := NewAdminService(psql.NewAdminRepo(database))

	// create services used in this test suite
	suite.residentService = NewResidentService(psql.NewResidentRepo(database))
	suite.authService = NewAuthService(jwtService, adminService, suite.residentService, config.HttpConfig{}, config.OAuthConfig{})

	// create resident
	suite.resident = models.Resident{
		ID:        "B1234567",
		FirstName: "Daniel",
		LastName:  "Velasquez",
		Phone:     "1234567890",
		Email:     "email@example.com",
		Password:  "notapassword"}
	if _, err := suite.residentService.Create(suite.resident); err != nil {
		suite.TearDownSuite()
		suite.T().Fatalf("tearing down because failed to create resident: %v", err)
	}
}

func (suite *authTestSuite) TearDownSuite() {
	err := suite.container.Terminate(context.Background())
	if err != nil {
		require.NoError(suite.T(), fmt.Errorf("error tearing down container: %v", err))
	}
}

func (suite *authTestSuite) TearDownTest() {
	if err := suite.residentService.residentRepo.Reset(); err != nil {
		suite.T().Fatalf("encountered error resetting auth repo in-between tests")
	}
}

func (suite *authTestSuite) TestResetPassword() {
	if err := suite.authService.ResetPassword(suite.resident.ID, "newPass"); err != nil {
		require.NoError(suite.T(), fmt.Errorf("error resetting password: %v", err))
	}
}
