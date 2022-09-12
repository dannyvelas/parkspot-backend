package api

import (
	"encoding/json"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type visitorRouterSuite struct {
	suite.Suite
	testServer    *httptest.Server
	residentJWT   string
	adminJWT      string
	testVisitor   newVisitorReq
	testVisitorId string
}

func TestVisitorRouter(t *testing.T) {
	suite.Run(t, new(visitorRouterSuite))
}

func (suite *visitorRouterSuite) SetupSuite() {
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

	{ // set jwts
		jwtMiddleware := NewJWTMiddleware(c.Token())

		suite.residentJWT, err = jwtMiddleware.newAccess(testResident.Id, ResidentRole)
		if err != nil {
			log.Fatal().Msgf("Failed to create JWT: %v", err)
		}

		suite.adminJWT, err = jwtMiddleware.newAccess("some-uuid", AdminRole)
		if err != nil {
			log.Fatal().Msgf("Failed to create JWT: %v", err)
		}
	}

	createTestResidents(suite.testServer.URL, suite.adminJWT)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	now := time.Now()
	suite.testVisitor = newVisitorReq{
		FirstName:    "Test",
		LastName:     "Visitor",
		Relationship: "fam/fri",
		AccessStart:  time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local),
		AccessEnd:    time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.Local),
	}

	suite.testVisitorId, err = createTestVisitor(suite.testServer.URL, suite.residentJWT, suite.testVisitor)
	if err != nil {
		log.Fatal().Msg("visitor_router_test.SetupSuite: Failed to create test visitor: " +
			err.Error())
	}
}

func (suite visitorRouterSuite) TearDownSuite() {
	defer suite.testServer.Close()

	// visitors that were created in `SetupSuite` would automatically be deleted
	// on CASCADE by the `deleteTestResidents` call below. So, technically this line isn't
	// necessary. But we include it anyway to ensure that visitor deletion works
	err := deleteTestVisitor(suite.testServer.URL, suite.residentJWT, suite.testVisitorId)
	if err != nil {
		log.Error().Msg("visitor_router_test.TearDownSuite: Failed to delete test visitor: " + err.Error())
		return
	}

	err = deleteTestResidents(suite.testServer.URL, suite.adminJWT)
	if err != nil {
		log.Error().Msg("visitor_router_test.TearDownSuite: " + err.Error())
		return
	}
}

func (suite visitorRouterSuite) TestGet_VisitorsOfResident_Positive() {
	responseBody, statusCode, err := authenticatedReq("GET", suite.testServer.URL+"/api/visitors", nil, suite.residentJWT)
	if err != nil {
		suite.NoError(fmt.Errorf("Error making request: %v", err))
		return
	}
	defer responseBody.Close()

	if statusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(responseBody)
		if err != nil {
			suite.NoError(fmt.Errorf("Error getting error response: %v", err))
			return
		}
		suite.Empty(string(bodyBytes))
		return
	}

	var response listWithMetadata[models.Visitor]
	if err := json.NewDecoder(responseBody).Decode(&response); err != nil {
		suite.NoError(fmt.Errorf("Error decoding response: %v", err))
		return
	}

	if len(response.Records) == 0 {
		suite.NotEmpty(response)
		return
	}

	firstVisitor := response.Records[0]
	suite.Equal(suite.testVisitorId, firstVisitor.Id)
	suite.Equal(suite.testVisitor.FirstName, firstVisitor.FirstName)
	suite.Equal(suite.testVisitor.LastName, firstVisitor.LastName)
}

func createTestVisitor(url string, jwt string, testVisitor newVisitorReq) (string, error) {
	requestBody, err := json.Marshal(testVisitor)
	if err != nil {
		return "", fmt.Errorf("Error marshalling testVisitor")
	}

	responseBody, statusCode, err := authenticatedReq("POST", url+"/api/visitor", requestBody, jwt)
	if err != nil {
		return "", fmt.Errorf("Error making request: %v", err)
	}
	defer responseBody.Close()

	if statusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(responseBody)
		if err != nil {
			return "", fmt.Errorf("Error getting error response: %v", err)
		}
		return "", fmt.Errorf("Bad response: %s", string(bodyBytes))
	}

	var response models.Visitor
	if err := json.NewDecoder(responseBody).Decode(&response); err != nil {
		return "", fmt.Errorf("Error decoding response: %v", err)
	}

	return response.Id, nil
}

func deleteTestVisitor(url string, jwt string, id string) error {
	endpoint := fmt.Sprintf("%s/api/visitor/%s", url, id)
	responseBody, statusCode, err := authenticatedReq("DELETE", endpoint, nil, jwt)
	if err != nil {
		return fmt.Errorf("Error making request: %v", err)
	}
	defer responseBody.Close()

	if statusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(responseBody)
		if err != nil {
			return fmt.Errorf("Error getting error response: %v", err)
		}
		return fmt.Errorf("Bad response: %s", string(bodyBytes))
	}

	return nil
}
