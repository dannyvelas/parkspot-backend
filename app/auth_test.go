package app

import (
	"context"
	"fmt"
	"github.com/dannyvelas/parkspot-backend/config"
	"github.com/dannyvelas/parkspot-backend/models"
	"github.com/dannyvelas/parkspot-backend/storage/psql"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

type authTestSuite struct {
	suite.Suite
	container       testcontainers.Container
	authService     AuthService
	residentService ResidentService // kept here so we can tear down between tests
}

func TestAuthService(t *testing.T) {
	suite.Run(t, new(authTestSuite))
}

func (suite *authTestSuite) SetupSuite() {
	// configure and start container
	container, database, err := psql.NewSandboxDatabase()
	if err != nil {
		suite.T().Fatalf("error getting sandbox database: %v", err)
	}
	// save container in suite struct so we can terminate it on suite teardown
	suite.container = container

	// create services that are needed to create services used in this suite
	jwtService := NewJWTService(config.TokenConfig{AccessSecret: "accessSecret", RefreshSecret: "refreshSecret"})
	adminService := NewAdminService(database.AdminRepo())

	// create services used in this test suite
	suite.residentService = NewResidentService(database.ResidentRepo())
	suite.authService = NewAuthService(jwtService, adminService, suite.residentService, config.HttpConfig{}, config.OAuthConfig{})

	// create resident
	if _, err := suite.residentService.Create(models.Test_resident); err != nil {
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
	const desiredPassword = "newPass"
	if err := suite.authService.ResetPassword(models.Test_resident.ID, desiredPassword); err != nil {
		require.NoError(suite.T(), fmt.Errorf("error resetting password: %v", err))
	}

	resident, err := suite.residentService.GetOne(models.Test_resident.ID)
	if err != nil {
		require.NoError(suite.T(), fmt.Errorf("error getting resident from database: %v", err))
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(resident.Password),
		[]byte(desiredPassword),
	); err != nil {
		require.NoError(suite.T(), fmt.Errorf("expected passwords to be the same but they werent: %v", err))
	}
}
