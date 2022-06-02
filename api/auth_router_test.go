package api

import (
	"bytes"
	"encoding/json"
	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
)

type authRouterSuite struct {
	suite.Suite
	testServer *httptest.Server
}

func TestAuthRouter(t *testing.T) {
	suite.Run(t, new(authRouterSuite))
}

func (suite *authRouterSuite) SetupSuite() {
	testServer, err := newTestServer()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	suite.testServer = testServer
}

func (suite authRouterSuite) TearDownSuite() {
	suite.testServer.Close()
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

	var userResponse user
	if err := json.NewDecoder(response.Body).Decode(&userResponse); err != nil {
		suite.NoError(err)
		return
	}

	expectedUser := newUser("b1394468-0018-47f5-afe5-1cc77118d161", "Daniel", "Velasquez", "email@example.com", AdminRole)

	suite.Equal(http.StatusOK, response.StatusCode)
	suite.Empty(cmp.Diff(expectedUser, userResponse), "response body was not the same")
}

func (suite authRouterSuite) TestLogin_Resident_Positive() {
	requestBody := []byte(`{
    "id":"T1043321",
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

	var userResponse user
	if err := json.NewDecoder(response.Body).Decode(&userResponse); err != nil {
		suite.NoError(err)
		return
	}

	expectedUser := newUser("T1043321", "John", "Gibson", "john.gibson@gmail.com", ResidentRole)

	suite.Equal(http.StatusOK, response.StatusCode)
	suite.Empty(cmp.Diff(expectedUser, userResponse), "response body was not the same")
}
