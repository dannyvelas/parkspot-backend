package api

import (
	"bytes"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type permitRouterSuite struct {
	suite.Suite
	testServer  *httptest.Server
	jwtToken    string
	existingCar newCarReq
}

func TestPermitRouter(t *testing.T) {
	suite.Run(t, new(permitRouterSuite))
}

func (suite *permitRouterSuite) SetupSuite() {
	config := config.NewConfig()

	suite.testServer = func() *httptest.Server {
		database, err := storage.NewDatabase(config.Postgres())
		if err != nil {
			log.Fatal().Msgf("Failed to start database: %v", err)
		}
		log.Info().Msg("Connected to Database.")

		// init repos
		adminRepo := storage.NewAdminRepo(database)
		permitRepo := storage.NewPermitRepo(database)
		carRepo := storage.NewCarRepo(database)
		residentRepo := storage.NewResidentRepo(database)

		// http setup
		httpConfig := config.Http()

		router := NewRouter(httpConfig, config.Token(), config.Constants().DateFormat(),
			adminRepo, permitRepo, carRepo, residentRepo)

		testServer := httptest.NewServer(router)
		log.Info().Msgf("Server started on: %s", testServer.URL)

		return testServer
	}()

	suite.jwtToken = func() string {
		jwtMiddleware := NewJWTMiddleware(config.Token())

		jwtToken, err := jwtMiddleware.newJWT("B1234567", "Daniel", "Velasquez", "example@email.com", AdminRole)
		if err != nil {
			log.Fatal().Msgf("Failed to create JWT token: %v", err)
		}

		return jwtToken
	}()

	suite.existingCar = newCarReq{"OGYR3X", "blue", "", ""}
}

func (suite permitRouterSuite) TearDownSuite() {
	suite.testServer.Close()
}

func (suite permitRouterSuite) TestCreate_NoStartNoEnd_Negative() {
	requestBody := []byte(`{
    "residentId":"T1043321",
    "car": {
      "licensePlate":"OGYR3X",
      "color":"blue",
      "make":"",
      "model":""
    }
  }`)
	request, err := http.NewRequest("POST", suite.testServer.URL+"/api/permit", bytes.NewBuffer(requestBody))
	if err != nil {
		suite.NoError(err)
		return
	}

	cookie := http.Cookie{Name: "jwt", Value: suite.jwtToken, HttpOnly: true, Path: "/"}
	request.AddCookie(&cookie)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		suite.NoError(err)
		return
	}
	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		suite.NoError(err)
		return
	}

	suite.Equal(http.StatusBadRequest, response.StatusCode)

	responseMsg := fmt.Sprintf("\"%v: startDate, endDate\"\n", errEmptyFields)
	suite.Equal(responseMsg, string(bodyBytes))
}

func (suite permitRouterSuite) TestCreate_EmptyStartEmptyEnd_Negative() {
	requestBody := []byte(`{
    "residentId":"T1043321",
    "car": {
      "licensePlate":"OGYR3X",
      "color":"blue",
      "make":"",
      "model":""
    },
    "startDate": "",
    "endDate": ""
  }`)
	request, err := http.NewRequest("POST", suite.testServer.URL+"/api/permit", bytes.NewBuffer(requestBody))
	if err != nil {
		suite.NoError(err)
		return
	}

	cookie := http.Cookie{Name: "jwt", Value: suite.jwtToken, HttpOnly: true, Path: "/"}
	request.AddCookie(&cookie)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		suite.NoError(err)
		return
	}
	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		suite.NoError(err)
		return
	}

	suite.Equal(http.StatusBadRequest, response.StatusCode)

	responseMsg := fmt.Sprintf("\"%v\"\n", newErrMalformed("NewPermitReq"))
	suite.Equal(responseMsg, string(bodyBytes))
}
