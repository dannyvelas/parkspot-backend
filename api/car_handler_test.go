package api

import (
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/app"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
)

type carRouterSuite struct {
	suite.Suite
	app        app.App
	testServer *httptest.Server
	adminJWT   string
}

func TestCarRouter(t *testing.T) {
	suite.Run(t, new(carRouterSuite))
}

func (suite *carRouterSuite) SetupSuite() {
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

func (suite carRouterSuite) TearDownSuite() {
	defer suite.testServer.Close()
}

func (suite carRouterSuite) TestEdit_CarDNE_Negative() {
	request := models.Car{LicensePlate: "NEWLP"}

	endpoint := fmt.Sprintf("%s/api/car/%s", suite.testServer.URL, "9b6d89a6-0b66-4170-be8d-eba43f8bf478")
	_, err := authenticatedReq[models.Car, models.Car]("PUT", endpoint, suite.adminJWT, &request)
	if err == nil {
		suite.NoError(fmt.Errorf("No error encountered when editing a non-existing car"))
		return
	}

	var apiErr *errs.ApiErr
	if !errors.As(err, &apiErr) {
		suite.NoError(fmt.Errorf("Couldn't cast error to apiErr. Error is: %v", err))
		return
	}

	suite.Equal(http.StatusNotFound, apiErr.StatusCode, "response was: %v", apiErr.Error())
}
