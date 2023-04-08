package app

import (
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/dannyvelas/lasvistas_api/util"
	"github.com/imdario/mergo"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

type residentTestSuite struct {
	suite.Suite
	residentService ResidentService
}

func TestResidentService(t *testing.T) {
	suite.Run(t, new(residentTestSuite))
}

func (suite *residentTestSuite) SetupSuite() {
	residentRepo := storage.NewResidentRepoMock()
	suite.residentService = NewResidentService(&residentRepo)
}

func (suite *residentTestSuite) TearDownTest() {
	suite.residentService.residentRepo.Reset()
}

func (suite residentTestSuite) TestCreate_ResidentDuplicateEmail_Negative() {
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

	err := suite.residentService.Create(resident1)
	if err != nil {
		suite.NoError(fmt.Errorf("Error creating resident when setting up test"))
		return
	}

	err = suite.residentService.Create(residentSameEmail)
	if err == nil {
		suite.NoError(fmt.Errorf("Successfully created resident with duplicate email when it shouldn't have"))
		return
	}

	var apiErr *errs.ApiErr
	if !errors.As(err, &apiErr) {
		suite.NoError(fmt.Errorf("Unexpected error: %v", err))
		return
	}

	suite.Equal(http.StatusBadRequest, apiErr.StatusCode, "Expected bad request got: %d", apiErr.StatusCode)
	suite.Contains(apiErr.Error(), "email") // assert bad request happened bc of email
}

func (suite residentTestSuite) TestEdit_Resident_Positive() {
	residentToEdit := models.Resident{
		ID:        "B0000000",
		FirstName: "first",
		LastName:  "last",
		Phone:     "1234567890",
		Email:     "email@example.com",
		Password:  "notapassword"}

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
			suite.NoError(fmt.Errorf("error merging when preparing test: %v\n", err))
			return
		}
		tests[testName] = test{argument: test_.argument, expected: expected}
	}

	// this func will execute one row of above table
	executeTest := func(test test) error {
		result, err := suite.residentService.Update(test.argument)
		if err != nil {
			return fmt.Errorf("Error making request: %v", err)
		}

		suite.Equal(test.expected.ID, result.ID)
		suite.Equal(test.expected.FirstName, result.FirstName)
		suite.Equal(test.expected.LastName, result.LastName)
		suite.Equal(test.expected.Phone, result.Phone)
		suite.Equal(test.expected.Email, result.Email)
		suite.Empty(result.Password) // residentResp.Password should be "" and not equal to test.expected.Password

		return nil
	}

	for testName, test := range tests {
		if err := suite.residentService.Create(residentToEdit); err != nil {
			suite.NoError(fmt.Errorf("Error creating test resident before running test: %v", err))
			break
		}

		if err := executeTest(test); err != nil {
			suite.NoError(fmt.Errorf("%s failed: %v", testName, err))
		}

		if err := suite.residentService.Delete(residentToEdit.ID); err != nil {
			suite.NoError(fmt.Errorf("Error deleting test resident after running test: %v", err))
			break
		}
	}
}

func (suite residentTestSuite) TestEdit_ResidentDNE_Negative() {
	residentThatDNE := models.Resident{ID: "B0000000", FirstName: "NEWFIRST"}

	_, err := suite.residentService.Update(residentThatDNE)
	if err == nil {
		suite.NoError(fmt.Errorf("No error encountered when editing a non-existing resident"))
		return
	}

	var apiErr *errs.ApiErr
	if !errors.As(err, &apiErr) {
		suite.NoError(fmt.Errorf("Couldn't cast error to apiErr. Error is: %v", err))
		return
	}

	suite.Equal(http.StatusNotFound, apiErr.StatusCode)
}
