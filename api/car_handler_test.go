package api

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/dannyvelas/parkspot-backend/app"
	"github.com/dannyvelas/parkspot-backend/config"
	"github.com/dannyvelas/parkspot-backend/models"
	"github.com/dannyvelas/parkspot-backend/storage/psql"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type carRouterSuite struct {
	suite.Suite
	container  testcontainers.Container
	testServer *httptest.Server
	app        app.App
}

func TestCarRouter(t *testing.T) {
	suite.Run(t, new(carRouterSuite))
}

func (suite *carRouterSuite) SetupSuite() {
	c, err := config.NewConfig()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	// configure and start container
	container, database, err := psql.NewSandboxDatabase()
	if err != nil {
		suite.T().Fatalf("error getting sandbox database: %v", err)
	}
	// save container in suite struct so we can terminate it on suite teardown
	suite.container = container

	suite.app = app.NewApp(c, database)

	router := newRouter(c, suite.app)
	suite.testServer = httptest.NewServer(router)

	// owner of car must exist before creating test car
	if _, err := suite.app.ResidentService.Create(models.TestResident); err != nil {
		log.Fatal().Msgf("error creating test resident: %v", err.Error())
	}
}

func (suite *carRouterSuite) TearDownSuite() {
	defer suite.testServer.Close()

	if err := suite.container.Terminate(context.Background()); err != nil {
		require.NoError(suite.T(), fmt.Errorf("error tearing down container: %v", err))
	}
}

func (suite *carRouterSuite) SetupTest() {
	// create fresh instance of car before each test
	if _, err := suite.app.CarService.Create(models.TestCar); err != nil {
		suite.TearDownSuite()
		suite.T().Fatalf("tearing down because failed to create resident: %v", err)
	}
}

func (suite *carRouterSuite) TearDownTest() {
	// delete car after each test
	if err := suite.app.CarService.Delete(models.TestCar.ID); err != nil {
		suite.TearDownSuite()
		suite.T().Fatalf("tearing down because failed to create resident: %v", err)
	}
}

func (suite *carRouterSuite) TestAdmin_Edit_Positive() {
	newColor := models.TestCar.Color + "NEW"

	token, err := suite.app.JWTService.NewAccess(models.TestAdmin.ID, models.AdminRole)
	if err != nil {
		require.NoError(suite.T(), fmt.Errorf("error creating access token for admin: %v", err))
	}

	returnedCar, _ := authenticatedReq[models.Car, models.Car]("PUT", suite.testServer.URL+"/api/car", token, &models.Car{
		ID:    models.TestCar.ID,
		Color: newColor,
	})

	expectedCar := models.TestCar
	expectedCar.Color = newColor
	require.Equal(suite.T(), expectedCar.ID, returnedCar.ID, "id in car response was not the same as expected")
	require.Equal(suite.T(), expectedCar.ResidentID, returnedCar.ResidentID, "residentID in car response was not the same as expected")
	require.Equal(suite.T(), expectedCar.LicensePlate, returnedCar.LicensePlate, "licensePlate in car response was not the same as expected")
	require.Equal(suite.T(), expectedCar.Color, returnedCar.Color, "color in car response was not the same as expected")
	require.Equal(suite.T(), expectedCar.Make, returnedCar.Make, "make in car response was not the same as expected")
	require.Equal(suite.T(), expectedCar.Model, returnedCar.Model, "model in car response was not the same as expected")
}

func (suite *carRouterSuite) TestSecurity_Edit_Negative() {
	newColor := models.TestCar.Color + "NEW"

	token, err := suite.app.JWTService.NewAccess(models.TestSecurity.ID, models.SecurityRole)
	if err != nil {
		require.NoError(suite.T(), fmt.Errorf("error creating access token for security: %v", err))
	}

	_, err = authenticatedReq[models.Car, models.Car]("PUT", suite.testServer.URL+"/api/car", token, &models.Car{
		ID:    models.TestCar.ID,
		Color: newColor,
	})
	require.Error(suite.T(), err)

	require.Contains(suite.T(), err.Error(), "unauthorized")
}

func (suite *carRouterSuite) TestResident_EditCar_Positive() {
	newColor := models.TestCar.Color + "NEW"

	token, err := suite.app.JWTService.NewAccess(models.TestResident.ID, models.ResidentRole)
	if err != nil {
		require.NoError(suite.T(), fmt.Errorf("error creating access token for resident: %v", err))
	}

	returnedCar, _ := authenticatedReq[models.Car, models.Car]("PUT", suite.testServer.URL+"/api/car", token, &models.Car{
		ID:    models.TestCar.ID,
		Color: newColor,
	})

	expectedCar := models.TestCar
	expectedCar.Color = newColor
	require.Equal(suite.T(), expectedCar.ID, returnedCar.ID, "id in car response was not the same as expected")
	require.Equal(suite.T(), expectedCar.ResidentID, returnedCar.ResidentID, "residentID in car response was not the same as expected")
	require.Equal(suite.T(), expectedCar.LicensePlate, returnedCar.LicensePlate, "licensePlate in car response was not the same as expected")
	require.Equal(suite.T(), expectedCar.Color, returnedCar.Color, "color in car response was not the same as expected")
	require.Equal(suite.T(), expectedCar.Make, returnedCar.Make, "make in car response was not the same as expected")
	require.Equal(suite.T(), expectedCar.Model, returnedCar.Model, "model in car response was not the same as expected")
}

func (suite *carRouterSuite) TestResident_EditOthersCar_Negative() {
	newColor := models.TestCar.Color + "NEW"

	// this is an access token belonging to models.TestResidentUnlimDays.
	// however, the car that is edited in the request belongs to models.TestResident
	token, err := suite.app.JWTService.NewAccess(models.TestResidentUnlimDays.ID, models.ResidentRole)
	if err != nil {
		require.NoError(suite.T(), fmt.Errorf("error creating access token for resident: %v", err))
	}

	_, err = authenticatedReq[models.Car, models.Car]("PUT", suite.testServer.URL+"/api/car", token, &models.Car{
		ID:    models.TestCar.ID,
		Color: newColor,
	})
	require.Error(suite.T(), err)

	require.Contains(suite.T(), err.Error(), "unauthorized")
}

func (suite *carRouterSuite) TestResident_DeleteOthersCar_Negative() {
	// this is an access token belonging to models.TestResidentUnlimDays.
	// however, the car that is deleted in the request belongs to models.TestResident
	token, err := suite.app.JWTService.NewAccess(models.TestResidentUnlimDays.ID, models.ResidentRole)
	if err != nil {
		require.NoError(suite.T(), fmt.Errorf("error creating access token for resident: %v", err))
	}

	endpoint := fmt.Sprintf("%s/api/car/%s", suite.testServer.URL, models.TestCar.ID)
	_, err = authenticatedReq[models.Car, models.Car]("DELETE", endpoint, token, nil)
	require.Error(suite.T(), err)

	require.Contains(suite.T(), err.Error(), "unauthorized")
}
