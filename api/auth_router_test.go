package api

import (
	"bytes"
	"encoding/json"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/storage"
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
}

func TestAuthRouter(t *testing.T) {
	suite.Run(t, new(authRouterSuite))
}

func (suite *authRouterSuite) SetupSuite() {
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

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		suite.NoError(err)
		return
	}

	expectedResponse, err := compressJSON([]byte(`{
    "id": "cca1e5d1-065b-47eb-98e7-731bfecd7a24",
    "firstName": "Daniel",
    "lastName": "Velasquez",
    "email": "email@example.com",
    "role": "admin"
  }`))
	if err != nil {
		suite.NoError(err)
		return
	}

	actualResponse, err := compressJSON(bodyBytes)
	if err != nil {
		suite.NoError(err)
		return
	}

	suite.Equal(http.StatusOK, response.StatusCode)
	suite.Empty(cmp.Diff(expectedResponse, actualResponse), "response body was not the same")
}

func compressJSON(jsonb []byte) ([]byte, error) {
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, jsonb); err != nil {
		return []byte{}, err
	}

	return buffer.Bytes(), nil
}
