package api

import (
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/app"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type permitRouterSuite struct {
	suite.Suite
	testServer     *httptest.Server
	app            app.App
	residentJWT    string
	adminJWT       string
	createdCar     models.Car
	testPermitReqs map[string]models.Permit
}

func TestPermitRouter(t *testing.T) {
	suite.Run(t, new(permitRouterSuite))
}

func (suite *permitRouterSuite) SetupSuite() {
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
			log.Fatal().Msgf("Failed to create resident JWT: %v", err)
		}

		suite.adminJWT, err = app.JWTService.NewAccess("some-uuid", models.AdminRole)
		if err != nil {
			log.Fatal().Msgf("Failed to create admin JWT: %v", err)
		}
	}

	err = createTestResidents(suite.app.ResidentService)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	suite.createdCar, err = app.CarService.Create(models.NewCar("id", "licensePlate", "color", "make", "model", 0))
	if err != nil {
		log.Fatal().Msgf("error creating car: %v", err)
	}

	suite.testPermitReqs = map[string]models.Permit{
		"NoUnlimDays,NoException": newTestPermitReq(testResident.ID, suite.createdCar.ID, ""),
		"UnlimDays,NoException":   newTestPermitReq(testResidentUnlimDays.ID, suite.createdCar.ID, ""),
		"NoUnlimDays,Exception":   newTestPermitReq(testResident.ID, suite.createdCar.ID, "some exception reason"),
		"UnlimDays,Exception":     newTestPermitReq(testResidentUnlimDays.ID, suite.createdCar.ID, "some exception reason"),
	}
}

func (suite permitRouterSuite) TearDownSuite() {
	defer suite.testServer.Close()

	err := deleteTestResidents(suite.app.ResidentService)
	if err != nil {
		log.Error().Msg("auth_router_test.TearDownSuite: " + err.Error())
		return
	}
}

func (suite permitRouterSuite) TestCreate_ResidentAndCarMultipleActivePermits() {
	// initialize and create a permit
	var permitOne = newTestPermitReq(testResident.ID, suite.createdCar.ID, "")
	{
		_, err := suite.app.PermitService.ValidateAndCreate(permitOne)
		if err != nil {
			suite.NoError(fmt.Errorf("Error creating permit: %v", err))
			return
		}
	}

	// initalize resident permits, each w a different car to eachother and to permitOne
	var resPermitTwo, resPermitThree models.Permit
	{
		carTwo, carTwoErr := suite.app.CarService.Create(models.NewCar("two", "two", "two", "two", "two", 0))
		carThree, carThreeErr := suite.app.CarService.Create(models.NewCar("three", "three", "three", "three", "three", 0))
		if carTwoErr != nil || carThreeErr != nil {
			suite.NoError(fmt.Errorf("Error creating carTwo: %v. or carThree: %v", carTwoErr, carThreeErr))
			return
		}
		resPermitTwo = newTestPermitReq(testResident.ID, carTwo.ID, "")
		resPermitThree = newTestPermitReq(testResident.ID, carThree.ID, "")
	}

	// initalize car permit, with the same car as permitOne
	carPermitTwo := newTestPermitReq(testResident.ID, permitOne.CarID, "")

	type createPermitTest struct {
		name       string
		permit     models.Permit
		shouldBeOk bool
	}

	// create an array of tests
	// each test is an array of permits to create
	type testSet []createPermitTest
	testSets := []testSet{
		{{"resident second permit", resPermitTwo, true}, {"resident third permit", resPermitThree, false}},
		{{"car second permit", carPermitTwo, false}},
	}

	// see which permit creations succeed/fail
	for _, testSet := range testSets {
		createdPermits := []models.Permit{}
		for _, createPermitTest := range testSet {
			permit, err := createTestPermit(suite.testServer.URL, suite.adminJWT, createPermitTest.permit)
			if err != nil && createPermitTest.shouldBeOk {
				suite.NoError(fmt.Errorf("%s failed: Error creating permit when it should've been okay: %v", createPermitTest.name, err))
			} else if err == nil && !createPermitTest.shouldBeOk {
				suite.NoError(fmt.Errorf("%s failed: Permit was created when it shouldn't have", createPermitTest.name))
			}
			// else: test passed: err == nil and should be okay OR err != nil and shouldn't be okay

			if err == nil {
				createdPermits = append(createdPermits, permit) // if permit was created, mark for deletion
			}
		}

		for _, permit := range createdPermits {
			err := deleteTestPermitAndCar(suite.app, suite.testServer.URL, suite.adminJWT, permit)
			if err != nil {
				suite.NoError(fmt.Errorf("Error deleting test permit and car: %v", err))
			}
		}
	}

	// delete permitOne and car associated with it
	err := deleteTestPermitAndCar(suite.app, suite.testServer.URL, suite.adminJWT, permitOne)
	if err != nil {
		err := fmt.Errorf("Error deleting test permit: %v", err)
		suite.NoError(err)
	}
}

func (suite permitRouterSuite) TestCreate_NoStartNoEnd_ErrMissing() {
	// initalize car with missing fields
	var carMissingFields models.Car
	{
		var err error
		carMissingFields, err = suite.app.CarService.Create(models.NewCar("id", "OGYR3X", "blue", "", "", 0))
		if err != nil {
			log.Error().Msg("auth_router_test.TearDownSuite: " + err.Error())
			return
		}
	}

	// initialize permit using car which is missing fields
	requestBody := newTestPermitReq(testResident.ID, carMissingFields.ID, "")

	_, err := authenticatedReq[models.Permit, models.Permit]("POST", suite.testServer.URL+"/api/permit", suite.adminJWT, &requestBody)
	if err == nil {
		suite.NoError(fmt.Errorf("Successfully created permit when it shouldn't have"))
		return
	}

	var resErr responseError
	if !errors.As(err, &resErr) {
		suite.NoError(fmt.Errorf("Unexpected error: %v", err))
		return
	}

	suite.Equal(http.StatusBadRequest, resErr.statusCode)

	responseMsg := fmt.Sprintf("\"%v: startDate, endDate\"\n", errEmptyFields)
	suite.Equal(responseMsg, resErr.message)
}

func (suite permitRouterSuite) TestCreate_EmptyStartEmptyEnd_ErrMalformed() {
	requestBody := models.Permit{
		ResidentID: testResident.ID,
		CarID:      suite.createdCar.ID,
	}
	_, err := authenticatedReq[models.Permit, models.Permit]("POST", suite.testServer.URL+"/api/permit", suite.adminJWT, &requestBody)
	if err == nil {
		suite.NoError(fmt.Errorf("Successfully created permit when it shouldn't have"))
		return
	}

	var resErr responseError
	if !errors.As(err, &resErr) {
		suite.NoError(fmt.Errorf("Unexpected error: %v", err))
		return
	}

	suite.Equal(http.StatusBadRequest, resErr.statusCode)

	responseMsg := fmt.Sprintf("\"%v\"\n", newErrMalformed("NewPermitReq"))
	suite.Equal(responseMsg, resErr.message)
}

func (suite permitRouterSuite) TestCreate_AddsResDays() {
	for testName, testPermitReq := range suite.testPermitReqs {
		residentBefore, err := suite.app.ResidentService.GetOne(testPermitReq.ResidentID)
		if err != nil {
			err := fmt.Errorf("%s failed: %v", testName, err)
			suite.NoError(err)
			return
		}

		createdPermit, err := createTestPermit(suite.testServer.URL, suite.adminJWT, testPermitReq)
		if err != nil {
			err := fmt.Errorf("%s failed: %v", testName, err)
			suite.NoError(err)
			return
		}

		residentNow, err := suite.app.ResidentService.GetOne(testPermitReq.ResidentID)
		if err != nil {
			err := fmt.Errorf("%s failed: %v", testName, err)
			suite.NoError(err)
			return
		}

		// shouldAddDays is true when permit residentID does not have unlim days
		// and exception reason is blank
		shouldAddDays := testPermitReq.ResidentID == testResident.ID && testPermitReq.ExceptionReason == ""

		lengthOfPermit := testPermitReq.EndDate.Sub(testPermitReq.StartDate)

		amtDaysAddedToRes := *residentNow.AmtParkingDaysUsed - *residentBefore.AmtParkingDaysUsed
		permitLength := int(lengthOfPermit.Hours() / 24)
		if !shouldAddDays && amtDaysAddedToRes != 0 {
			err := fmt.Errorf("%s failed: added %d days when it shouldn't have. Permit length was: %d", testName, amtDaysAddedToRes, permitLength)
			suite.NoError(err)
			return
		} else if shouldAddDays && amtDaysAddedToRes != permitLength {
			err := fmt.Errorf("%s failed: added incorrect amount of days: %d instead of %d", testName, amtDaysAddedToRes, permitLength)
			suite.NoError(err)
			return
		}

		err = deleteTestPermitAndCar(suite.app, suite.testServer.URL, suite.adminJWT, createdPermit)
		if err != nil {
			suite.NoError(fmt.Errorf("%s failed: deleting test permit failed: %v", testName, err))
			return
		}
	}
}

func (suite permitRouterSuite) TestDelete_SubtractsResDays() {
	for testName, testPermitReq := range suite.testPermitReqs {
		residentBefore, err := suite.app.ResidentService.GetOne(testPermitReq.ResidentID)
		if err != nil {
			suite.NoError(fmt.Errorf("%s failed: %v", testName, err))
			return
		}

		createdPermit, err := createTestPermit(suite.testServer.URL, suite.adminJWT, testPermitReq)
		if err != nil {
			suite.NoError(fmt.Errorf("%s failed: %v", testName, err))
			return
		}

		err = deleteTestPermitAndCar(suite.app, suite.testServer.URL, suite.adminJWT, createdPermit)
		if err != nil {
			suite.NoError(fmt.Errorf("%s failed: %v", testName, err))
			return
		}

		residentNow, err := suite.app.ResidentService.GetOne(testPermitReq.ResidentID)
		if err != nil {
			suite.NoError(fmt.Errorf("%s failed: %v", testName, err))
			return
		}

		if residentBefore.AmtParkingDaysUsed != residentNow.AmtParkingDaysUsed {
			suite.NoError(fmt.Errorf("%s failed: did not subract days. Resident has %d instead of %d", testName, residentNow.AmtParkingDaysUsed, residentBefore.AmtParkingDaysUsed))
			return
		}
	}
}

func (suite permitRouterSuite) TestDelete_AddsCarDays() {
}

func (suite permitRouterSuite) TestDelete_SubtractsCarDays() {
}

func (suite permitRouterSuite) TestCreate_AllFieldsMatch() {
}

func (suite permitRouterSuite) TestGetActivePermitsOfResident_Postive() {
	var createdPermit = newTestPermitReq(testResident.ID, suite.createdCar.ID, "")
	{
		_, err := suite.app.PermitService.ValidateAndCreate(createdPermit)
		if err != nil {
			suite.NoError(fmt.Errorf("Error creating permit: %v", err))
			return
		}
	}

	endpoint := fmt.Sprintf("%s/api/permits/active", suite.testServer.URL)
	permitsResponse, err := authenticatedReq[any, models.ListWithMetadata[models.Permit]]("GET", endpoint, suite.residentJWT, nil)
	if err != nil {
		suite.NoError(err)
		return
	}

	if len(permitsResponse.Records) == 0 {
		suite.NotEmpty(permitsResponse.Records, "length of permits should not be zero")
		return
	}

	last := permitsResponse.Records[len(permitsResponse.Records)-1]

	suite.Equal(createdPermit.ResidentID, last.ResidentID)
	suite.Equal(createdPermit.LicensePlate, last.LicensePlate)
	suite.Empty(cmp.Diff(createdPermit.StartDate, last.StartDate))
	suite.Empty(cmp.Diff(createdPermit.EndDate, last.EndDate))

	err = deleteTestPermitAndCar(suite.app, suite.testServer.URL, suite.adminJWT, createdPermit)
	if err != nil {
		suite.NoError(fmt.Errorf("Error deleting test permit and car: %v", err))
	}
}

func (suite permitRouterSuite) TestGetMaxExceptions_Positive() {
	endpoint := fmt.Sprintf("%s/api/permits/exceptions?limit=%d", suite.testServer.URL, config.MaxLimit)
	permitsResponse, err := authenticatedReq[any, models.ListWithMetadata[models.Permit]]("GET", endpoint, suite.adminJWT, nil)
	if err != nil {
		suite.NoError(err)
		return
	}

	if len(permitsResponse.Records) == 0 {
		suite.NotEmpty(permitsResponse.Records, "length of permits should not be zero")
		return
	}

	if permitsResponse.Metadata.TotalAmount < config.MaxLimit {
		suite.Equal(permitsResponse.Metadata.TotalAmount, len(permitsResponse.Records), "The amount of records reported in metadata is lower than limit, so the amount of records in the payload should be equal to metadata.totalAmount")
	}
}

// helpers
func newTestPermitReq(residentID, carID, exceptionReason string) models.Permit {
	return models.Permit{
		ResidentID:      residentID,
		CarID:           carID,
		StartDate:       time.Now().Truncate(time.Second),
		EndDate:         time.Now().Add(time.Hour * 24).Truncate(time.Second),
		ExceptionReason: exceptionReason,
	}
}

func createTestPermit(url string, adminJWT string, desiredPermit models.Permit) (models.Permit, error) {
	response, err := authenticatedReq[models.Permit, models.Permit]("POST", url+"/api/permit", adminJWT, &desiredPermit)
	if err != nil {
		return models.Permit{}, fmt.Errorf("Error making request: %v", err)
	}
	return response, nil
}

func deleteTestPermitAndCar(app app.App, url, adminJWT string, permit models.Permit) error {
	err := app.PermitService.Delete(permit.ID)
	if err != nil {
		return err
	}

	err = app.CarService.Delete(permit.CarID)
	if err != nil {
		return fmt.Errorf("Error deleting car directly in carRepo: %v", err)
	}

	return nil
}
