package api

import (
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/app"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/util"
	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var (
	testResidentUnlimDays = models.Resident{
		ID:        "B7654321",
		FirstName: "Daniel",
		LastName:  "Velasquez",
		Phone:     "1234567890",
		Email:     "email2@example.com",
		Password:  "notapassword",
		UnlimDays: util.ToPtr(true)}
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

	suite.app, err = app.NewApp(c)
	if err != nil {
		log.Fatal().Msgf("Failed to initialize app: %v", err)
	}

	router := newRouter(c, suite.app)
	suite.testServer = httptest.NewServer(router)

	{ // set jwts
		suite.residentJWT, err = suite.app.JWTService.NewAccess(testResident.ID, models.ResidentRole)
		if err != nil {
			log.Fatal().Msgf("Failed to create resident JWT: %v", err)
		}

		suite.adminJWT, err = suite.app.JWTService.NewAccess("some-uuid", models.AdminRole)
		if err != nil {
			log.Fatal().Msgf("Failed to create admin JWT: %v", err)
		}
	}

	if err := suite.app.ResidentService.Create(testResident); err != nil {
		log.Fatal().Msgf("error creating test resident: %v", err.Error())
	}

	if err := suite.app.ResidentService.Create(testResidentUnlimDays); err != nil {
		log.Fatal().Msgf("error creating test resident with unlimited days: %v", err.Error())
	}

	suite.createdCar, err = suite.app.CarService.Create(models.NewCar("id", testResident.ID, "lp1", "color", "make", "model", 0))
	if err != nil {
		log.Fatal().Msgf("error creating car: %v", err.Error())
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

	if err := suite.app.ResidentService.Delete(testResident.ID); err != nil {
		log.Fatal().Msgf("error deleting test resident: %v", err.Error())
	}

	if err := suite.app.ResidentService.Delete(testResidentUnlimDays.ID); err != nil {
		log.Fatal().Msgf("error deleting test resident with unlimited days: %v", err.Error())
	}

	if err := suite.app.CarService.Delete(suite.createdCar.ID); err != nil {
		log.Fatal().Msgf("error deleting created car: %v", err.Error())
	}
}

func (suite permitRouterSuite) TestCreate_ResidentAndCarMultipleActivePermits() {
	// initialize and create a permit
	permitOne, err := suite.app.PermitService.ValidateAndCreate(newTestPermitReq(testResident.ID, suite.createdCar.ID, ""))
	if err != nil {
		suite.NoError(fmt.Errorf("Error creating permit: %v", err))
		return
	}

	// initalize resident permits, each w a different car to eachother and to permitOne
	var resPermitTwo, resPermitThree models.Permit
	{
		carTwo, carTwoErr := suite.app.CarService.Create(models.NewCar("two", testResident.ID, "two", "two", "two", "two", 0))
		carThree, carThreeErr := suite.app.CarService.Create(models.NewCar("three", testResident.ID, "three", "three", "three", "three", 0))
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
			permit, err := authenticatedReq[models.Permit, models.Permit]("POST", suite.testServer.URL+"/api/permit", suite.adminJWT, &createPermitTest.permit)
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
			err := suite.deletePermitAndCar(permit)
			if err != nil {
				suite.NoError(fmt.Errorf("Error deleting test permit and car: %v", err))
			}
		}
	}

	// delete permitOne
	err = suite.app.PermitService.Delete(permitOne.ID)
	if err != nil {
		err := fmt.Errorf("Error deleting test permit: %v", err)
		suite.NoError(err)
	}
}

func (suite permitRouterSuite) TestCreate_NoStartNoEnd_ErrMissing() {
	// initalize car with missing fields
	createdCar, err := suite.app.CarService.Create(models.NewCar("id", testResident.ID, "lp2", "color", "make", "model", 0))
	if err != nil {
		suite.NoError(fmt.Errorf("Error creating car with missing fields: %v", err))
		return
	}

	// initialize permit using car above
	requestBody := models.Permit{
		ResidentID:      testResident.ID,
		CarID:           createdCar.ID,
		ExceptionReason: "",
	}

	_, err = authenticatedReq[models.Permit, models.Permit]("POST", suite.testServer.URL+"/api/permit", suite.adminJWT, &requestBody)
	if err == nil {
		suite.NoError(fmt.Errorf("Successfully created permit when it shouldn't have"))
		return
	}

	var apiErr *errs.ApiErr
	if !errors.As(err, &apiErr) {
		suite.NoError(fmt.Errorf("Unexpected error: %v", err))
		return
	}

	suite.Equal(http.StatusBadRequest, apiErr.StatusCode)
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

	var apiErr *errs.ApiErr
	if !errors.As(err, &apiErr) {
		suite.NoError(fmt.Errorf("Unexpected error: %v", err))
		return
	}

	suite.Equal(http.StatusBadRequest, apiErr.StatusCode)
}

func (suite permitRouterSuite) TestCreate_AddsResDays() {
	for testName, testPermitReq := range suite.testPermitReqs {
		residentBefore, err := suite.app.ResidentService.GetOne(testPermitReq.ResidentID)
		if err != nil {
			suite.NoError(fmt.Errorf("%s failed: %v", testName, err))
			return
		}

		createdPermit, err := authenticatedReq[models.Permit, models.Permit]("POST", suite.testServer.URL+"/api/permit", suite.adminJWT, &testPermitReq)
		if err != nil {
			suite.NoError(fmt.Errorf("%s failed: %v", testName, err))
			return
		}

		residentNow, err := suite.app.ResidentService.GetOne(testPermitReq.ResidentID)
		if err != nil {
			suite.NoError(fmt.Errorf("%s failed: %v", testName, err))
			return
		}

		// shouldAddDays is true when permit residentID does not have unlim days
		// and exception reason is blank
		shouldAddDays := testPermitReq.ResidentID == testResident.ID && testPermitReq.ExceptionReason == ""

		amtDaysAddedToRes := *residentNow.AmtParkingDaysUsed - *residentBefore.AmtParkingDaysUsed
		permitLength := util.GetAmtDays(testPermitReq.StartDate, testPermitReq.EndDate)
		if !shouldAddDays && amtDaysAddedToRes != 0 {
			err := fmt.Errorf("%s failed: added %d days when it shouldn't have. Permit length was: %d", testName, amtDaysAddedToRes, permitLength)
			suite.NoError(err)
			return
		} else if shouldAddDays && amtDaysAddedToRes != permitLength {
			err := fmt.Errorf("%s failed: added incorrect amount of days: %d instead of %d", testName, amtDaysAddedToRes, permitLength)
			suite.NoError(err)
			return
		}

		err = suite.app.PermitService.Delete(createdPermit.ID)
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

		createdPermit, err := authenticatedReq[models.Permit, models.Permit]("POST", suite.testServer.URL+"/api/permit", suite.adminJWT, &testPermitReq)
		if err != nil {
			suite.NoError(fmt.Errorf("%s failed: %v", testName, err))
			return
		}

		err = suite.app.PermitService.Delete(createdPermit.ID)
		if err != nil {
			suite.NoError(fmt.Errorf("%s failed: %v", testName, err))
			return
		}

		residentNow, err := suite.app.ResidentService.GetOne(testPermitReq.ResidentID)
		if err != nil {
			suite.NoError(fmt.Errorf("%s failed: %v", testName, err))
			return
		}

		if *residentBefore.AmtParkingDaysUsed != *residentNow.AmtParkingDaysUsed {
			suite.NoError(fmt.Errorf("%s failed: did not subract days. Resident has %d instead of %d", testName, *residentNow.AmtParkingDaysUsed, *residentBefore.AmtParkingDaysUsed))
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
	createdPermit, err := suite.app.PermitService.ValidateAndCreate(newTestPermitReq(testResident.ID, suite.createdCar.ID, ""))
	if err != nil {
		suite.NoError(fmt.Errorf("Error creating permit: %v", err))
		return
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

	err = suite.app.PermitService.Delete(createdPermit.ID)
	if err != nil {
		suite.NoError(fmt.Errorf("Error deleting test permit and car: %v", err))
	}
}

func (suite permitRouterSuite) TestGetMaxExceptions_Positive() {
	createdPermit, err := suite.app.PermitService.ValidateAndCreate(newTestPermitReq(testResident.ID, suite.createdCar.ID, "an exception reason here"))
	if err != nil {
		suite.NoError(fmt.Errorf("error creating test permit: %v", err))
		return
	}

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

	last := permitsResponse.Records[len(permitsResponse.Records)-1]

	suite.Equal(createdPermit.ResidentID, last.ResidentID)
	suite.Equal(createdPermit.LicensePlate, last.LicensePlate)
	suite.Empty(cmp.Diff(createdPermit.StartDate, last.StartDate))
	suite.Empty(cmp.Diff(createdPermit.EndDate, last.EndDate))

	err = suite.app.PermitService.Delete(createdPermit.ID)
	if err != nil {
		suite.NoError(fmt.Errorf("error creating test permit: %v", err))
		return
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

func (suite permitRouterSuite) deletePermitAndCar(permit models.Permit) error {
	if err := suite.app.PermitService.Delete(permit.ID); err != nil {
		return err
	}

	if err := suite.app.CarService.Delete(permit.CarID); err != nil {
		return fmt.Errorf("Error deleting car directly in carRepo: %v", err)
	}

	return nil
}
