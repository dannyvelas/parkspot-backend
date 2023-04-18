package app

import (
	"context"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage/psql"
	"github.com/imdario/mergo"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"net/http"
	"testing"
)

type carTestSuite struct {
	suite.Suite
	container  testcontainers.Container
	carService CarService
	resident   models.Resident // will exist in db for duration of tests
}

func TestCarService(t *testing.T) {
	suite.Run(t, new(carTestSuite))
}

func (suite *carTestSuite) SetupSuite() {
	// configure and start container
	container, database, err := getSandboxDatabase()
	if err != nil {
		suite.T().Fatalf("error getting sandbox database: %v", err)
	}
	// save container in suite struct so we can terminate it on suite teardown
	suite.container = container

	suite.resident = models.Resident{
		ID:        "B1234567",
		FirstName: "Daniel",
		LastName:  "Velasquez",
		Phone:     "1234567890",
		Email:     "email@example.com",
		Password:  "notapassword"}
	residentService := NewResidentService(psql.NewResidentRepo(database))
	if _, err := residentService.Create(suite.resident); err != nil {
		suite.TearDownSuite()
		suite.T().Fatalf("tearing down because failed to create resident: %v", err)
	}

	carRepo := psql.NewCarRepo(database)
	suite.carService = NewCarService(carRepo)
}

func (suite carTestSuite) TearDownSuite() {
	err := suite.container.Terminate(context.Background())
	if err != nil {
		require.NoError(suite.T(), fmt.Errorf("error tearing down container: %v", err))
	}
}

func (suite carTestSuite) TearDownTest() {
	err := suite.carService.carRepo.Reset()
	if err != nil {
		suite.T().Fatalf("encountered error resetting car repo in-between tests")
	}
}

func (suite carTestSuite) TestEdit_CarDNE_Negative() {
	carWithIDThatDNE := models.Car{ID: "9b6d89a6-0b66-4170-be8d-eba43f8bf478", LicensePlate: "NEWLP"}

	_, err := suite.carService.Update(carWithIDThatDNE)
	require.Error(suite.T(), err, "No error encountered when editing a non-existing car")

	var apiErr *errs.ApiErr
	require.ErrorAsf(suite.T(), err, &apiErr, "Couldn't cast error to apiErr. Error is: %v", err)

	require.Equal(suite.T(), http.StatusNotFound, apiErr.StatusCode, "response was: %v", apiErr.Error())
}

func (suite carTestSuite) TestEdit_Car_Positive() {
	carToEdit := models.NewCar("d1e0affb-14e7-4e9f-b8a3-70be7d49d063", suite.resident.ID, "lp1", "color", "make", "model", 0)

	// set up a table of tests
	type test struct {
		argument models.Car
		expected models.Car
	}
	tests := map[string]test{
		"licensePlate":              {argument: models.Car{ID: carToEdit.ID, LicensePlate: "NEWLP"}},
		"licensePlate, color":       {argument: models.Car{ID: carToEdit.ID, LicensePlate: "NEWLP", Color: "NEWCOLOR"}},
		"licensePlate, color, make": {argument: models.Car{ID: carToEdit.ID, LicensePlate: "NEWLP", Color: "NEWCOLOR", Make: "NEWMAKE"}},
		"model":                     {argument: models.Car{ID: carToEdit.ID, Model: "NEWMODEL"}},
	}
	for testName, test_ := range tests {
		expected := test_.argument
		if err := mergo.Merge(&expected, carToEdit); err != nil {
			require.NoError(suite.T(), fmt.Errorf("error merging when preparing test: %v\n", err))
		}
		tests[testName] = test{argument: test_.argument, expected: expected}
	}

	// this func will execute one row of above table
	executeTest := func(test test) error {
		result, err := suite.carService.Update(test.argument)
		if err != nil {
			return fmt.Errorf("Error making request: %v", err)
		}

		suite.Equal(test.expected.ID, result.ID)
		suite.Equal(test.expected.LicensePlate, result.LicensePlate)
		suite.Equal(test.expected.Color, result.Color)
		suite.Equal(test.expected.Make, result.Make)
		suite.Equal(test.expected.Model, result.Model)

		return nil
	}

	for testName, test := range tests {
		createdCar, err := suite.carService.Create(carToEdit)
		require.NoError(suite.T(), err)

		err = executeTest(test)
		if err != nil {
			require.NoError(suite.T(), fmt.Errorf("%s failed: %v", testName, err))
		}

		err = suite.carService.Delete(createdCar.ID)
		if err != nil {
			require.NoError(suite.T(), err)
		}
	}
}

func (suite carTestSuite) TestCreate_CarRepeatLP_Negative() {
	prevExistingCar := models.NewCar("", suite.resident.ID, "lp1", "color", "make", "model", 0)
	if _, err := suite.carService.Create(prevExistingCar); err != nil {
		require.NoError(suite.T(), fmt.Errorf("Error creating test car before running test: %v", err))
	}

	carWithSameLP := models.NewCar("", suite.resident.ID, "lp1", "color", "make", "model", 0)
	_, err := suite.carService.Create(carWithSameLP)
	require.NotNil(suite.T(), err, "error when creating car with duplicate LP was not nil but it should have been")

	require.ErrorIs(suite.T(), err, errs.AlreadyExists, "error is expected to be one of already exists")
}
