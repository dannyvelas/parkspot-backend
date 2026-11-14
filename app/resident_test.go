package app

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/dannyvelas/parkspot-backend/errs"
	"github.com/dannyvelas/parkspot-backend/models"
	"github.com/dannyvelas/parkspot-backend/storage"
	"github.com/dannyvelas/parkspot-backend/util"
	"github.com/imdario/mergo"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type residentTestSuite struct {
	suite.Suite
	residentService ResidentService
}

func TestResidentService(t *testing.T) {
	suite.Run(t, new(residentTestSuite))
}

func (suite *residentTestSuite) SetupTest() {
	residentRepo := storage.NewResidentRepoMock()
	suite.residentService = NewResidentService(&residentRepo)
}

func (suite *residentTestSuite) TestCreate_ResidentDuplicateEmail_Negative() {
	resident1 := models.Resident{
		ID:        "B0000000",
		FirstName: "first",
		LastName:  "resident",
		Phone:     "123456789",
		Email:     "email@example.com",
		Password:  "password",
		UnlimDays: util.ToPtr(false),
	}
	residentSameEmail := models.Resident{
		ID:        "B1111111",
		FirstName: "second",
		LastName:  "resident",
		Phone:     "123456789",
		Email:     "email@example.com",
		Password:  "password",
		UnlimDays: util.ToPtr(false),
	}

	_, err := suite.residentService.Create(resident1)
	require.NoError(suite.T(), err, "error creating resident when setting up test")

	_, err = suite.residentService.Create(residentSameEmail)
	require.NotNil(suite.T(), err, "successfully created resident with duplicate email when it shouldn't have")

	var apiErr *errs.APIErr
	require.ErrorAs(suite.T(), err, &apiErr, "unexpected error: %v", err)

	suite.Equal(http.StatusBadRequest, apiErr.StatusCode, "Expected bad request got: %d", apiErr.StatusCode)
	suite.Contains(apiErr.Error(), "email") // assert bad request happened bc of email
}

func (suite *residentTestSuite) TestEdit_Resident_Positive() {
	residentToEdit := models.Resident{
		ID:        "B0000000",
		FirstName: "first",
		LastName:  "last",
		Phone:     "1234567890",
		Email:     "email@example.com",
		Password:  "notapassword",
	}

	// set up a table of tests
	type test struct {
		argument models.Resident
		expected models.Resident
	}
	tests := map[string]test{
		"firstName":                  {argument: models.Resident{ID: residentToEdit.ID, FirstName: "NEWFIRST"}},
		"firstName, lastName":        {argument: models.Resident{ID: residentToEdit.ID, FirstName: "NEWFIRST", LastName: "NEWLAST"}},
		"firstName, lastName, phone": {argument: models.Resident{ID: residentToEdit.ID, FirstName: "NEWFIRST", LastName: "NEWLAST", Phone: "06181999"}},
		"unlimDays":                  {argument: models.Resident{ID: residentToEdit.ID, UnlimDays: util.ToPtr(true)}},
		"amtParkingDaysUsed":         {argument: models.Resident{ID: residentToEdit.ID, AmtParkingDaysUsed: util.ToPtr(42)}},
	}
	for testName, test_ := range tests {
		expected := test_.argument
		if err := mergo.Merge(&expected, residentToEdit); err != nil {
			require.NoError(suite.T(), fmt.Errorf("error merging when preparing test: %v\n", err))
		}
		tests[testName] = test{argument: test_.argument, expected: expected}
	}

	// this func will execute one row of above table
	executeTest := func(test test) error {
		result, err := suite.residentService.Update(test.argument)
		if err != nil {
			return fmt.Errorf("error making request: %v", err)
		}

		suite.Equal(test.expected.ID, result.ID)
		suite.Equal(test.expected.FirstName, result.FirstName)
		suite.Equal(test.expected.LastName, result.LastName)
		suite.Equal(test.expected.Phone, result.Phone)
		suite.Equal(test.expected.Email, result.Email)
		// passwords changes will be checked in another test, they require slightly different logic

		return nil
	}

	for testName, test := range tests {
		_, err := suite.residentService.Create(residentToEdit)
		require.NoError(suite.T(), err, "error creating test resident before running test")

		if err := executeTest(test); err != nil {
			require.NoError(suite.T(), fmt.Errorf("%s failed: %v", testName, err))
		}

		err = suite.residentService.Delete(residentToEdit.ID)
		require.NoError(suite.T(), err, "error deleting test resident after running test")
	}
}

func (suite *residentTestSuite) TestEdit_ResidentDNE_Negative() {
	residentThatDNE := models.Resident{ID: "B0000000", FirstName: "NEWFIRST"}

	_, err := suite.residentService.Update(residentThatDNE)
	require.NotNil(suite.T(), err, "no error encountered when editing a non-existing resident")

	var apiErr *errs.APIErr
	require.ErrorAs(suite.T(), err, &apiErr, "couldn't cast error to apiErr. Error is: %v", err)

	suite.Equal(http.StatusNotFound, apiErr.StatusCode)
}
