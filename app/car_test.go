package app

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/imdario/mergo"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

type carTestSuite struct {
	suite.Suite
	carService CarService
}

func TestCarService(t *testing.T) {
	suite.Run(t, new(carTestSuite))
}

func (suite *carTestSuite) SetupSuite() {
	carRepo := storage.NewCarRepoMock()
	suite.carService = NewCarService(&carRepo)
}

func (suite *carTestSuite) TearDownTest() {
	suite.carService.carRepo.Reset()
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
	carToEdit := models.NewCar("d1e0affb-14e7-4e9f-b8a3-70be7d49d063", "B0000000", "lp1", "color", "make", "model", 0)

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
			suite.NoError(fmt.Errorf("error merging when preparing test: %v\n", err))
			return
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
		require.NoError(suite.T(), err, fmt.Errorf("Error creating test car before running test: %v", err))

		err = executeTest(test)
		require.NoError(suite.T(), err, fmt.Errorf("%s failed: %v", testName, err))

		err = suite.carService.Delete(createdCar.ID)
		require.NoError(suite.T(), err, fmt.Errorf("Error deleting test car after running test: %v", err))
	}
}

func (suite carTestSuite) TestCreate_CarRepeatLP_Negative() {
	prevExistingCar := models.NewCar("", "B0000000", "lp1", "color", "make", "model", 0)
	_, err := suite.carService.Create(prevExistingCar)
	require.NoError(suite.T(), err, fmt.Errorf("Error creating test car before running test: %v", err))

	carWithSameLP := models.NewCar("", "B0000000", "lp1", "color", "make", "model", 0)
	_, err = suite.carService.Create(carWithSameLP)
	require.NotNil(suite.T(), err, "error when creating car with duplicate LP was not nil but it should have been")

	require.ErrorIs(suite.T(), err, errs.AlreadyExists, "error is expected to be one of already exists")
}
