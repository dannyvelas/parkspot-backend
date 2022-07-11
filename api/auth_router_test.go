package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type authRouterSuite struct {
	suite.Suite
	testServer *httptest.Server
	jwtToken   string
}

func TestAuthRouter(t *testing.T) {
	suite.Run(t, new(authRouterSuite))
}

func (suite *authRouterSuite) SetupSuite() {
	config, err := config.NewConfig()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	suite.testServer, err = newTestServer()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	suite.jwtToken, err = getJWTToken(config.Token())
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	err = createTestResidents(suite.testServer.URL, suite.jwtToken)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
}

func (suite authRouterSuite) TearDownSuite() {
	defer suite.testServer.Close()

	err := deleteTestResidents(suite.testServer.URL, suite.jwtToken)
	if err != nil {
		log.Error().Msg("auth_router_test.TearDownSuite: " + err.Error())
		return
	}
}

func (suite authRouterSuite) TestLogin_Admin_Positive() {
	requestBody := []byte(`{
    "id":"email@example.com",
    "password":"notapassword"
  }`)
	request, err := http.NewRequest("POST", suite.testServer.URL+"/api/login", bytes.NewBuffer(requestBody))
	if err != nil {
		suite.NoError(err)
		return
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		suite.NoError(err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			suite.NoError(err)
			return
		}
		suite.Empty(string(bodyBytes))
		return
	}

	var userResponse user
	if err := json.NewDecoder(response.Body).Decode(&userResponse); err != nil {
		suite.NoError(err)
		return
	}

	expectedUser := newUser("b1394468-0018-47f5-afe5-1cc77118d161",
		"Daniel",
		"Velasquez",
		"email@example.com",
		AdminRole)
	suite.Empty(cmp.Diff(expectedUser, userResponse), "response body was not the same")
}

func (suite authRouterSuite) TestLogin_Resident_Positive() {
	requestBody := []byte(fmt.Sprintf(`{
    "id":"%s",
    "password":"%s"
  }`, testResident.Id, testResident.Password))
	request, err := http.NewRequest("POST", suite.testServer.URL+"/api/login", bytes.NewBuffer(requestBody))
	if err != nil {
		suite.NoError(err)
		return
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		suite.NoError(err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			suite.NoError(err)
			return
		}
		suite.Empty(string(bodyBytes))
		return
	}

	var userResponse user
	if err := json.NewDecoder(response.Body).Decode(&userResponse); err != nil {
		suite.NoError(err)
		return
	}

	expectedUser := newUser(testResident.Id,
		testResident.FirstName,
		testResident.LastName,
		testResident.Email,
		ResidentRole)
	suite.Empty(cmp.Diff(expectedUser, userResponse), "response body was not the same")
}
