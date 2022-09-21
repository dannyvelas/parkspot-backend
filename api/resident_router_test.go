package api

import (
	"encoding/json"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
	"net/http/httptest"
	"testing"
)

type residentRouterSuite struct {
	suite.Suite
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

	database, err := storage.NewDatabase(c.Postgres())
	if err != nil {
		log.Fatal().Msgf("Failed to start database: %v", err)
	}

	repos := storage.NewRepos(database)

	suite.testServer = newTestServer(c, repos)

	jwtMiddleware := NewJWTMiddleware(c.Token())

	suite.adminJWT, err = jwtMiddleware.newAccess("some-uuid", AdminRole)
	if err != nil {
		log.Fatal().Msgf("Failed to create JWT: %v", err)
	}

	err = createTestResidents(suite.testServer.URL, suite.adminJWT)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
}

func (suite residentRouterSuite) TearDownSuite() {
	defer suite.testServer.Close()

	err := deleteTestResidents(suite.testServer.URL, suite.adminJWT)
	if err != nil {
		log.Error().Msg("resident_router_test.TearDownSuite: " + err.Error())
		return
	}
}

func (suite residentRouterSuite) TestEdit_Resident_Positive() {
	requestBody := []byte(`{"firstName":"NEWFIRSTNAME"}`)
	endpoint := fmt.Sprintf("%s/api/resident/%s", suite.testServer.URL, testResident.Id)
	responseBody, err := authenticatedReq("PUT", endpoint, requestBody, suite.adminJWT)
	if err != nil {
		suite.NoError(fmt.Errorf("Error making request: %v", err))
		return
	}
	defer responseBody.Close()

	var actualResident models.Resident
	if err := json.NewDecoder(responseBody).Decode(&actualResident); err != nil {
		suite.NoError(err)
		return
	}

	expectedResident := testResident
	expectedResident.FirstName = "NEWFIRSTNAME"
	expectedResident.Password = "" // passwords are not included in JSON responses

	suite.Empty(cmp.Diff(expectedResident, actualResident), "user in response did not equal expected user")
}
