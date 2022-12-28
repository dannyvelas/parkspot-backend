package api

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/app"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/util"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
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

	app, err := app.NewApp(c)
	if err != nil {
		log.Fatal().Msgf("Failed to initialize app: %v", err)
	}
	suite.app = app

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
			models.Resident{FirstName: "NEWFIRST"},
			merge(testResident, models.Resident{FirstName: "NEWFIRST"}),
		},
		"firstName, lastName": {
			models.Resident{FirstName: "NEWFIRST", LastName: "NEWLAST"},
			merge(testResident, models.Resident{FirstName: "NEWFIRST", LastName: "NEWLAST"}),
		},
		"firstName, lastName, phone": {
			models.Resident{FirstName: "NEWFIRST", LastName: "NEWLAST", Phone: "06181999"},
			merge(testResident, models.Resident{FirstName: "NEWFIRST", LastName: "NEWLAST", Phone: "06181999"}),
		},
		"unlimDays": {
			models.Resident{UnlimDays: util.ToPtr(true)},
			merge(testResident, models.Resident{UnlimDays: util.ToPtr(true)}),
		},
		"amtParkingDaysUsed": {
			models.Resident{AmtParkingDaysUsed: util.ToPtr(42)},
			merge(testResident, models.Resident{AmtParkingDaysUsed: util.ToPtr(42)}),
		},
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
		suite.Equal(test.expected.Password, residentResp.Password)

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

func merge(res1, res2 models.Resident) models.Resident {
	mergeResult := res1

	if res2.FirstName != "" {
		mergeResult.FirstName = res2.FirstName
	}
	if res2.LastName != "" {
		mergeResult.LastName = res2.LastName
	}
	if res2.Phone != "" {
		mergeResult.Phone = res2.Phone
	}
	if res2.Email != "" {
		mergeResult.Email = res2.Email
	}
	if res2.UnlimDays != nil {
		mergeResult.UnlimDays = res2.UnlimDays
	}
	if res2.AmtParkingDaysUsed != nil {
		mergeResult.AmtParkingDaysUsed = res2.AmtParkingDaysUsed
	}

	mergeResult.Password = "" // passwords are always "" in JSON responses

	return mergeResult
}
