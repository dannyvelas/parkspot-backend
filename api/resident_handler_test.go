package api

import (
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/app"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/util"
	"github.com/imdario/mergo"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
)

type residentRouterSuite struct {
	suite.Suite
	app        app.App
	testServer *httptest.Server
	adminJWT   string
}

func TestResidentRouter(t *testing.T) {
	suite.Run(t, new(residentRouterSuite))
}

func (suite *residentRouterSuite) SetupSuite() {
	c, err := config.NewConfig()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	suite.app, err = app.NewApp(c)
	if err != nil {
		log.Fatal().Msgf("Failed to initialize app: %v", err)
	}

	router := newRouter(c, suite.app)
	suite.testServer = httptest.NewServer(router)

	suite.adminJWT, err = suite.app.JWTService.NewAccess("some-uuid", models.AdminRole)
	if err != nil {
		log.Fatal().Msgf("Failed to create JWT: %v", err)
	}
}

func (suite residentRouterSuite) TearDownSuite() {
	defer suite.testServer.Close()
}

func (suite residentRouterSuite) TestEdit_Resident_Positive() {
	type test struct {
		request  models.Resident
		expected models.Resident
	}

	tests := map[string]test{
		"firstName": {
			request: models.Resident{FirstName: "NEWFIRST"},
		},
		"firstName, lastName": {
			request: models.Resident{FirstName: "NEWFIRST", LastName: "NEWLAST"},
		},
		"firstName, lastName, phone": {
			request: models.Resident{FirstName: "NEWFIRST", LastName: "NEWLAST", Phone: "06181999"},
		},
		"unlimDays": {
			request: models.Resident{UnlimDays: util.ToPtr(true)},
		},
		"amtParkingDaysUsed": {
			request: models.Resident{AmtParkingDaysUsed: util.ToPtr(42)},
		},
	}
	for testName, test_ := range tests {
		expected := test_.request
		if err := mergo.Merge(&expected, testResident); err != nil {
			suite.NoError(fmt.Errorf("error merging when preparing test: %v\n", err))
			return
		}
		tests[testName] = test{request: test_.request, expected: expected}
	}

	executeTest := func(test test) error {
		endpoint := fmt.Sprintf("%s/api/resident/%s", suite.testServer.URL, testResident.ID)
		residentResp, err := authenticatedReq[models.Resident, models.Resident]("PUT", endpoint, suite.adminJWT, &test.request)
		if err != nil {
			return fmt.Errorf("Error making request: %v", err)
		}

		suite.Equal(test.expected.ID, residentResp.ID)
		suite.Equal(test.expected.FirstName, residentResp.FirstName)
		suite.Equal(test.expected.LastName, residentResp.LastName)
		suite.Equal(test.expected.Phone, residentResp.Phone)
		suite.Equal(test.expected.Email, residentResp.Email)
		suite.Empty(residentResp.Password) // residentResp.Password should be "" and not equal to test.expected.Password

		return nil
	}

	for testName, test := range tests {
		if err := suite.app.ResidentService.Create(testResident); err != nil {
			suite.NoError(fmt.Errorf("Error creating test resident before running test: %v", err))
			break
		}

		if err := executeTest(test); err != nil {
			suite.NoError(fmt.Errorf("%s failed: %v", testName, err))
		}

		if err := suite.app.ResidentService.Delete(testResident.ID); err != nil {
			suite.NoError(fmt.Errorf("Error deleting test resident after running test: %v", err))
			break
		}
	}
}

func (suite residentRouterSuite) TestEdit_ResidentDNE_Negative() {
	request := models.Resident{FirstName: "NEWFIRST"}

	endpoint := fmt.Sprintf("%s/api/resident/%s", suite.testServer.URL, testResident.ID)
	_, err := authenticatedReq[models.Resident, models.Resident]("PUT", endpoint, suite.adminJWT, &request)
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
