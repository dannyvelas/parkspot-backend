package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
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
	carRepo     storage.CarRepo
	testServer  *httptest.Server
	residentJWT string
	adminJWT    string
	testPermits map[string]newPermitReq
	testPermit  newPermitReq // noUnlimDays,noException
}

func TestPermitRouter(t *testing.T) {
	suite.Run(t, new(permitRouterSuite))
}

func (suite *permitRouterSuite) SetupSuite() {
	c, err := config.NewConfig()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	database, err := storage.NewDatabase(c.Postgres())
	if err != nil {
		log.Fatal().Msgf("Failed to start database: %v", err)
	}

	repos := storage.NewRepos(database)

	suite.carRepo = repos.Car

	suite.testServer = newTestServer(c, repos)

	{ // set jwts
		jwtMiddleware := NewJWTMiddleware(c.Token())

		suite.residentJWT, err = jwtMiddleware.newAccess(testResident.ID, ResidentRole)
		if err != nil {
			log.Fatal().Msgf("Failed to create JWT: %v", err)
		}

		suite.adminJWT, err = jwtMiddleware.newAccess("some-uuid", AdminRole)
		if err != nil {
			log.Fatal().Msgf("Failed to create JWT: %v", err)
		}
	}

	err = createTestResidents(suite.testServer.URL, suite.adminJWT)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	carOneReq := newCarReq{"one", "one", "one", "one"}
	suite.testPermits = map[string]newPermitReq{
		"NoUnlimDays,NoException": newTestPermit(testResident.ID, carOneReq, ""),
		"UnlimDays,NoException":   newTestPermit(testResidentUnlimDays.ID, carOneReq, ""),
		"NoUnlimDays,Exception":   newTestPermit(testResident.ID, carOneReq, "some exception reason"),
		"UnlimDays,Exception":     newTestPermit(testResidentUnlimDays.ID, carOneReq, "some exception reason"),
	}
	suite.testPermit = newTestPermit(testResident.ID, carOneReq, "")
}

func (suite permitRouterSuite) TearDownSuite() {
	defer suite.testServer.Close()

	err := deleteTestResidents(suite.testServer.URL, suite.adminJWT)
	if err != nil {
		log.Error().Msg("permit_router_test.TearDownSuite: " + err.Error())
		return
	}
}

func (suite permitRouterSuite) TestCreate_ResidentAndCarMultipleActivePermits() {
	permitOne := suite.testPermit

	// resident permits, each w a different car to eachother and to permitOne
	resPermitTwo := newTestPermit(testResident.ID, newCarReq{"two", "two", "two", "two"}, "")
	resPermitThree := newTestPermit(testResident.ID, newCarReq{"three", "three", "three", "three"}, "")

	// car permit, with the same car as permitOne
	carPermitTwo := newTestPermit(testResident.ID, permitOne.Car, "")

	type createPermitTest struct {
		name       string
		permit     newPermitReq
		shouldBeOk bool
	}

	// create an array of tests
	// each test is an array of permits to create
	type testSet []createPermitTest
	testSets := []testSet{
		{{"resident second permit", resPermitTwo, true}, {"resident third permit", resPermitThree, false}},
		{{"car second permit", carPermitTwo, false}},
	}

	// create permitOne
	createdPermitOne, err := createTestPermit(suite.testServer.URL, suite.adminJWT, permitOne)
	if err != nil {
		suite.NoError(fmt.Errorf("Error creating permitOne: %v", err))
		return
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
			err = deleteTestPermitAndCar(suite.testServer.URL, suite.adminJWT, permit.ID, permit.Car.ID, suite.carRepo)
			if err != nil {
				suite.NoError(fmt.Errorf("Error deleting test permit and car: %v", err))
			}
		}
	}

	// delete permitOne and car associated with it
	err = deleteTestPermitAndCar(suite.testServer.URL, suite.adminJWT, createdPermitOne.ID, createdPermitOne.Car.ID, suite.carRepo)
	if err != nil {
		err := fmt.Errorf("Error deleting test permit: %v", err)
		suite.NoError(err)
	}
}

func (suite permitRouterSuite) TestCreate_FillInCarFields() {
	// car that will be seeded in db with missing fields
	seedCar_licensePlate := "two"
	seedCar_color := "two"
	seedCar_make := ""
	seedCar_model := ""

	// permit that will be created after that car is seeded
	// lp is the same, so `carToCreate` will be used as the permit's car
	newCarReq := newCarReq{LicensePlate: "two", Color: "two", Make: "two", Model: "two"}
	permitToCreate := newTestPermit(testResident.ID, newCarReq, "")

	carID, err := suite.carRepo.Create(seedCar_licensePlate, seedCar_color, seedCar_make, seedCar_model)
	if err != nil {
		err := fmt.Errorf("Error creating car directly in carRepo: %v", err)
		suite.NoError(err)
		return
	}

	createdPermit, err := createTestPermit(suite.testServer.URL, suite.adminJWT, permitToCreate)
	if err != nil {
		err := fmt.Errorf("Error creating test permit: %v", err)
		suite.NoError(err)
		return
	}

	carAfter, err := suite.carRepo.GetOne(carID)
	if err != nil {
		err := fmt.Errorf("Error getting car directly from carRepo: %v", err)
		suite.NoError(err)
		return
	}

	err = deleteTestPermitAndCar(suite.testServer.URL, suite.adminJWT, createdPermit.ID, createdPermit.Car.ID, suite.carRepo)
	if err != nil {
		err := fmt.Errorf("Error deleting test permit: %v", err)
		suite.NoError(err)
		return
	}

	suite.NotEmpty(carAfter.Make, "when permit was created, existing car did not get make field filled in")
	suite.NotEmpty(carAfter.Model, "when permit was created, existing car did not get model field filled in")
}

func (suite permitRouterSuite) TestCreate_NoStartNoEnd_ErrMissing() {
	requestBody := []byte(`{
    "residentID":"T1043321",
    "car": {
      "licensePlate":"OGYR3X",
      "color":"blue",
      "make":"",
      "model":""
    }
  }`)
	responseBody, err := authenticatedReq("POST", suite.testServer.URL+"/api/permit", requestBody, suite.adminJWT)
	if err == nil {
		defer responseBody.Close()
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
	requestBody := []byte(`{
    "residentID":"T1043321",
    "car": {
      "licensePlate":"OGYR3X",
      "color":"blue",
      "make":"",
      "model":""
    },
    "startDate": "",
    "endDate": ""
  }`)
	responseBody, err := authenticatedReq("POST", suite.testServer.URL+"/api/permit", requestBody, suite.adminJWT)
	if err == nil {
		defer responseBody.Close()
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
	for testName, testPermit := range suite.testPermits {
		residentBefore, err := getTestResident(suite.testServer.URL, testPermit.ResidentID, suite.adminJWT)
		if err != nil {
			err := fmt.Errorf("%s failed: %v", testName, err)
			suite.NoError(err)
			return
		}

		createdPermit, err := createTestPermit(suite.testServer.URL, suite.adminJWT, testPermit)
		if err != nil {
			err := fmt.Errorf("%s failed: %v", testName, err)
			suite.NoError(err)
			return
		}

		residentNow, err := getTestResident(suite.testServer.URL, testPermit.ResidentID, suite.adminJWT)
		if err != nil {
			err := fmt.Errorf("%s failed: %v", testName, err)
			suite.NoError(err)
			return
		}

		// shouldAddDays is true when permit residentID does not have unlim days
		// and exception reason is blank
		shouldAddDays := testPermit.ResidentID == testResident.ID && testPermit.ExceptionReason == ""

		lengthOfPermit := testPermit.EndDate.Sub(testPermit.StartDate)

		amtDaysAddedToRes := residentNow.AmtParkingDaysUsed - residentBefore.AmtParkingDaysUsed
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

		err = deleteTestPermitAndCar(suite.testServer.URL, suite.adminJWT, createdPermit.ID, createdPermit.Car.ID, suite.carRepo)
		if err != nil {
			suite.NoError(fmt.Errorf("%s failed: deleting test permit failed: %v", testName, err))
			return
		}
	}
}

func (suite permitRouterSuite) TestDelete_SubtractsResDays() {
	for testName, newPermitReq := range suite.testPermits {
		residentBefore, err := getTestResident(suite.testServer.URL, newPermitReq.ResidentID, suite.adminJWT)
		if err != nil {
			suite.NoError(fmt.Errorf("%s failed: %v", testName, err))
			return
		}

		createdPermit, err := createTestPermit(suite.testServer.URL, suite.adminJWT, newPermitReq)
		if err != nil {
			suite.NoError(fmt.Errorf("%s failed: %v", testName, err))
			return
		}

		err = deleteTestPermitAndCar(suite.testServer.URL, suite.adminJWT, createdPermit.ID, createdPermit.Car.ID, suite.carRepo)
		if err != nil {
			suite.NoError(fmt.Errorf("%s failed: %v", testName, err))
			return
		}

		residentNow, err := getTestResident(suite.testServer.URL, newPermitReq.ResidentID, suite.adminJWT)
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
	createdPermit, err := createTestPermit(suite.testServer.URL, suite.adminJWT, suite.testPermit)
	if err != nil {
		suite.NoError(fmt.Errorf("Error creating permit: %v", err))
		return
	}
	defer func() {
		err = deleteTestPermitAndCar(suite.testServer.URL, suite.adminJWT, createdPermit.ID, createdPermit.Car.ID, suite.carRepo)
		if err != nil {
			suite.NoError(fmt.Errorf("Error deleting test permit and car: %v", err))
		}
	}()

	endpoint := fmt.Sprintf("%s/api/permits/active", suite.testServer.URL)
	responseBody, err := authenticatedReq("GET", endpoint, nil, suite.residentJWT)
	if err != nil {
		suite.NoError(err)
		return
	}
	defer responseBody.Close()

	var permitsResponse listWithMetadata[models.Permit]
	if err := json.NewDecoder(responseBody).Decode(&permitsResponse); err != nil {
		suite.NoError(err)
		return
	} else if len(permitsResponse.Records) == 0 {
		suite.NotEmpty(permitsResponse.Records, "length of permits should not be zero")
		return
	}

	last := permitsResponse.Records[len(permitsResponse.Records)-1]
	suite.Equal(suite.testPermit.ResidentID, last.ResidentID)
	suite.Equal(suite.testPermit.Car.LicensePlate, last.Car.LicensePlate)
	suite.Empty(cmp.Diff(last.StartDate, suite.testPermit.StartDate))
	suite.Empty(cmp.Diff(last.EndDate, suite.testPermit.EndDate))
}

func (suite permitRouterSuite) TestGetMaxExceptions_Positive() {
	endpoint := fmt.Sprintf("%s/api/permits/exceptions?limit=%d", suite.testServer.URL, config.MaxLimit)
	responseBody, err := authenticatedReq("GET", endpoint, nil, suite.adminJWT)
	if err != nil {
		suite.NoError(err)
		return
	}
	defer responseBody.Close()

	var permitsResponse listWithMetadata[models.Permit]
	if err := json.NewDecoder(responseBody).Decode(&permitsResponse); err != nil {
		suite.NoError(err)
		return
	} else if len(permitsResponse.Records) == 0 {
		suite.NotEmpty(permitsResponse.Records, "length of permits should not be zero")
		return
	}

	if permitsResponse.Metadata.TotalAmount < config.MaxLimit {
		suite.Equal(permitsResponse.Metadata.TotalAmount, len(permitsResponse.Records), "The amount of records reported in metadata is lower than limit, so the amount of records in the payload should be equal to metadata.totalAmount")
	}
}

// helpers
func newTestPermit(residentID string, car newCarReq, exceptionReason string) newPermitReq {
	return newPermitReq{
		ResidentID:      residentID,
		Car:             car,
		StartDate:       time.Now().Truncate(time.Second),
		EndDate:         time.Now().Add(time.Duration(24) * time.Hour).Truncate(time.Second),
		ExceptionReason: exceptionReason,
	}
}

func createTestPermit(url string, adminJWT string, testPermit newPermitReq) (models.Permit, error) {
	requestBody, err := json.Marshal(testPermit)
	if err != nil {
		return models.Permit{}, fmt.Errorf("Error marshalling: %v", err)
	}

	responseBody, err := authenticatedReq("POST", url+"/api/permit", requestBody, adminJWT)
	if err != nil {
		return models.Permit{}, fmt.Errorf("Error making request: %v", err)
	}
	defer responseBody.Close()

	var newPermitResponse models.Permit
	if err := json.NewDecoder(responseBody).Decode(&newPermitResponse); err != nil {
		return models.Permit{}, fmt.Errorf("Error decoding response: %v", err)
	}

	return newPermitResponse, nil
}

func deleteTestPermitAndCar(url, adminJWT string, permitID int, carID string, carRepo storage.CarRepo) error {
	endpoint := fmt.Sprintf("%s/api/permit/%d", url, permitID)
	responseBody, err := authenticatedReq("DELETE", endpoint, nil, adminJWT)
	if err != nil {
		return fmt.Errorf("Error making request: %v", err)
	}
	defer responseBody.Close()

	err = carRepo.Delete(carID)
	if err != nil {
		return fmt.Errorf("Error deleting car directly in carRepo: %v", err)
	}

	return nil
}

func getTestResident(url string, residentID string, adminJWT string) (models.Resident, error) {
	endpoint := fmt.Sprintf("%s/api/resident/%s", url, residentID)
	responseBody, err := authenticatedReq("GET", endpoint, nil, adminJWT)
	if err != nil {
		return models.Resident{}, err
	}
	defer responseBody.Close()

	var resident models.Resident
	if err := json.NewDecoder(responseBody).Decode(&resident); err != nil {
		return models.Resident{}, err
	}

	return resident, nil
}
