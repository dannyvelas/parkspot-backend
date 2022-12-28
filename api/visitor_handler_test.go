package api

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/app"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
	"net/http/httptest"
	"testing"
	"time"
)

type visitorRouterSuite struct {
	suite.Suite
	app            app.App
	testServer     *httptest.Server
	residentJWT    string
	adminJWT       string
	createdVisitor models.Visitor
}

func TestVisitorRouter(t *testing.T) {
	suite.Run(t, new(visitorRouterSuite))
}

func (suite *visitorRouterSuite) SetupSuite() {
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

	{ // set jwts
		suite.residentJWT, err = app.JWTService.NewAccess(testResident.ID, models.ResidentRole)
		if err != nil {
			log.Fatal().Msgf("Failed to create JWT: %v", err)
		}

		suite.adminJWT, err = app.JWTService.NewAccess("some-uuid", models.AdminRole)
		if err != nil {
			log.Fatal().Msgf("Failed to create JWT: %v", err)
		}
	}

	if err := suite.app.ResidentService.Create(testResident); err != nil {
		log.Fatal().Msgf("error creating test resident: %v", err.Error())
	}

	now := time.Now()
	testVisitor := models.NewVisitor(
		"",
		testResident.ID,
		"Test",
		"Visitor",
		"fam/fri",
		time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local),
		time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.Local),
	)

	suite.createdVisitor, err = suite.app.VisitorService.Create(testVisitor)
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
	err := suite.app.VisitorService.Delete(suite.createdVisitor.ID)
	if err != nil {
		log.Fatal().Msgf("visitor_router_test.TearDownSuite: Failed to delete test visitor: " + err.Error())
	}

	if err := suite.app.ResidentService.Delete(testResident.ID); err != nil {
		log.Fatal().Msgf("error deleting test resident: %v", err.Error())
	}
}

func (suite visitorRouterSuite) TestGet_VisitorsOfResident_Positive() {
	visitorsResp, err := authenticatedReq[any, models.ListWithMetadata[models.Visitor]]("GET", suite.testServer.URL+"/api/visitors", suite.residentJWT, nil)
	if err != nil {
		suite.NoError(fmt.Errorf("Error making request: %v", err))
		return
	}

	if len(visitorsResp.Records) == 0 {
		suite.NotEmpty(visitorsResp)
		return
	}

	firstVisitor := visitorsResp.Records[0]
	suite.Equal(suite.createdVisitor.ID, firstVisitor.ID)
	suite.Equal(suite.createdVisitor.FirstName, firstVisitor.FirstName)
	suite.Equal(suite.createdVisitor.LastName, firstVisitor.LastName)
}
