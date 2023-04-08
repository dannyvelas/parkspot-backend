package app

import (
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/dannyvelas/lasvistas_api/util"
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
