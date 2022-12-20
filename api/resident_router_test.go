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

	// init database
	database, err := storage.NewDatabase(c.Postgres())
	if err != nil {
		log.Fatal().Msgf("Failed to start database: %v", err)
	}

	{ // init test server
		repos := storage.NewRepos(database)
		suite.testServer = newTestServer(c, repos)
	}

	{ // set jwts
		jwtMiddleware := NewJWTMiddleware(c.Token())
		suite.adminJWT, err = jwtMiddleware.newAccess("some-uuid", AdminRole)
		if err != nil {
			log.Fatal().Msgf("Failed to create JWT: %v", err)
		}
	}
}

func (suite residentRouterSuite) TearDownSuite() {
	defer suite.testServer.Close()
}

func (suite residentRouterSuite) TestEdit_Resident_Positive() {
	type test struct {
		request  string
		expected models.Resident
	}

	// used below
	boolTrue := true
	num42 := 42

	tests := map[string]test{
		"firstName": {
			`{"firstName":"NEWFIRST"}`,
			testResidentWith(editResidentReq{FirstName: "NEWFIRST"}),
		},
		"firstName, lastName": {
			`{"firstName":"NEWFIRST","lastName":"NEWLAST"}`,
			testResidentWith(editResidentReq{FirstName: "NEWFIRST", LastName: "NEWLAST"}),
		},
		"firstName, lastName, phone": {
			`{"firstName":"NEWFIRST","lastName":"NEWLAST","phone":"06181999"}`,
			testResidentWith(editResidentReq{FirstName: "NEWFIRST", LastName: "NEWLAST", Phone: "06181999"}),
		},
		"unlimDays": {
			`{"unlimDays":true}`,
			testResidentWith(editResidentReq{UnlimDays: &boolTrue}),
		},
		"amtParkingDaysUsed": {
			`{"amtParkingDaysUsed":42}`,
			testResidentWith(editResidentReq{AmtParkingDaysUsed: &num42}),
		},
	}

	executeTest := func(test test) error {
		requestBody := []byte(test.request)
		endpoint := fmt.Sprintf("%s/api/resident/%s", suite.testServer.URL, testResident.ID)
		responseBody, err := authenticatedReq("PUT", endpoint, requestBody, suite.adminJWT)
		if err != nil {
			return fmt.Errorf("Error making request: %v", err)
		}
		defer responseBody.Close()

		var actualResident models.Resident
		if err := json.NewDecoder(responseBody).Decode(&actualResident); err != nil {
			return err
		}

		if difference := cmp.Diff(test.expected, actualResident); difference != "" {
			return fmt.Errorf("user in response did not equal expected user: " + difference)
		}

		return nil
	}

	for testName, test := range tests {
		err := hitCreateResidentEndpoint(suite.testServer.URL, suite.adminJWT, testResident)
		if err != nil {
			suite.NoError(fmt.Errorf("Error creating test resident before running test: %v", err))
			break
		}

		err = executeTest(test)
		if err != nil {
			suite.NoError(fmt.Errorf("%s failed: %v", testName, err))
		}

		err = hitDeleteResidentEndpoint(suite.testServer.URL, suite.adminJWT, testResident.ID)
		if err != nil {
			suite.NoError(fmt.Errorf("Error deleting test resident after running test: %v", err))
			break
		}
	}
}

func testResidentWith(override editResidentReq) models.Resident {
	returnResident := testResident

	if override.FirstName != "" {
		returnResident.FirstName = override.FirstName
	}
	if override.LastName != "" {
		returnResident.LastName = override.LastName
	}
	if override.Phone != "" {
		returnResident.Phone = override.Phone
	}
	if override.Email != "" {
		returnResident.Email = override.Email
	}
	if override.UnlimDays != nil {
		returnResident.UnlimDays = *override.UnlimDays
	}
	if override.AmtParkingDaysUsed != nil {
		returnResident.AmtParkingDaysUsed = *override.AmtParkingDaysUsed
	}

	returnResident.Password = "" // passwords are always "" in JSON responses

	return returnResident
}
