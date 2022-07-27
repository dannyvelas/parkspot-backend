package api

import (
	"encoding/json"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type permitRouterSuite struct {
	suite.Suite
	carRepo     storage.CarRepo
	testServer  *httptest.Server
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

	suite.testServer = newTestServer(c, repos)

	suite.adminJWT, err = getAdminJWT(c.Token())
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	err = createTestResidents(suite.testServer.URL, suite.adminJWT)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	testCar := newCarReq{"one", "one", "one", "one"}
	suite.testPermits = map[string]newPermitReq{
		"NoUnlimDays,NoException": newTestPermit(testResident.Id, testCar, ""),
		"UnlimDays,NoException":   newTestPermit(testResidentUnlimDays.Id, testCar, ""),
		"NoUnlimDays,Exception":   newTestPermit(testResident.Id, testCar, "some exception reason"),
		"UnlimDays,Exception":     newTestPermit(testResidentUnlimDays.Id, testCar, "some exception reason"),
	}
	suite.testPermit = newTestPermit(testResident.Id, testCar, "")
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
	resPermitTwo := newTestPermit(testResident.Id, newCarReq{"two", "two", "two", "two"}, "")
	resPermitThree := newTestPermit(testResident.Id, newCarReq{"three", "three", "three", "three"}, "")

	// car permit, with the same car as permitOne
	carPermitTwo := newTestPermit(testResident.Id, permitOne.Car, "")

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
	id, err := createTestPermit(suite.testServer.URL, suite.adminJWT, permitOne)
	if err != nil {
		suite.NoError(err)
		return
	}
	defer deleteTestPermit(suite.testServer.URL, suite.adminJWT, id)

	// see which permit creations succeed/fail
	for _, testSet := range testSets {
		createdPermitIds := []int{}
		for _, createPermitTest := range testSet {
			id, err := createTestPermit(suite.testServer.URL, suite.adminJWT, createPermitTest.permit)
			if err != nil && createPermitTest.shouldBeOk {
				suite.NoError(fmt.Errorf("%s failed: Error creating permit when it should've been okay: %v", createPermitTest.name, err))
			} else if err == nil && !createPermitTest.shouldBeOk {
				suite.NoError(fmt.Errorf("%s failed: Permit was created when it shouldn't have", createPermitTest.name))
			}
			// else: test passed: err == nil and should be okay OR err != nil and shouldn't be okay

			if err == nil {
				createdPermitIds = append(createdPermitIds, id) // if permit was created, mark for deletion
			}
		}

		for _, id := range createdPermitIds {
			err = deleteTestPermit(suite.testServer.URL, suite.adminJWT, id)
			if err != nil {
				suite.NoError(fmt.Errorf("Error deleting test permit: %v", err))
			}
		}
	}
}

func (suite permitRouterSuite) TestCreate_NoStartNoEnd_ErrMissing() {
	requestBody := []byte(`{
    "residentId":"T1043321",
    "car": {
      "licensePlate":"OGYR3X",
      "color":"blue",
      "make":"",
      "model":""
    }
  }`)
	responseBody, statusCode, err := authenticatedReq("POST", suite.testServer.URL+"/api/permit", requestBody, suite.adminJWT)
	if err != nil {
		suite.NoError(err)
		return
	}
	defer responseBody.Close()

	bodyBytes, err := io.ReadAll(responseBody)
	if err != nil {
		suite.NoError(err)
		return
	}

	suite.Equal(http.StatusBadRequest, statusCode)

	responseMsg := fmt.Sprintf("\"%v: startDate, endDate\"\n", errEmptyFields)
	suite.Equal(responseMsg, string(bodyBytes))
}

func (suite permitRouterSuite) TestCreate_EmptyStartEmptyEnd_ErrMalformed() {
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
	responseBody, statusCode, err := authenticatedReq("POST", suite.testServer.URL+"/api/permit", requestBody, suite.adminJWT)
	if err != nil {
		suite.NoError(err)
		return
	}
	defer responseBody.Close()

	bodyBytes, err := io.ReadAll(responseBody)
	if err != nil {
		suite.NoError(err)
		return
	}

	suite.Equal(http.StatusBadRequest, statusCode)

	responseMsg := fmt.Sprintf("\"%v\"\n", newErrMalformed("NewPermitReq"))
	suite.Equal(responseMsg, string(bodyBytes))
}

func (suite permitRouterSuite) TestCreate_AddsResDays() {
	executeTest := func(testName string, testPermit newPermitReq) error {
		residentBefore, err := getTestResident(suite.testServer.URL, testPermit.ResidentId, suite.adminJWT)
		if err != nil {
			return fmt.Errorf("%s failed: %v", testName, err)
		}

		id, err := createTestPermit(suite.testServer.URL, suite.adminJWT, testPermit)
		if err != nil {
			return fmt.Errorf("%s failed: %v", testName, err)
		}
		defer deleteTestPermit(suite.testServer.URL, suite.adminJWT, id)

		residentNow, err := getTestResident(suite.testServer.URL, testPermit.ResidentId, suite.adminJWT)
		if err != nil {
			return fmt.Errorf("%s failed: %v", testName, err)
		}

		// shouldAddDays is true when permit residentId does not have unlim days
		// and exception reason is blank
		shouldAddDays := testPermit.ResidentId == testResident.Id && testPermit.ExceptionReason == ""

		lengthOfPermit := testPermit.EndDate.Sub(testPermit.StartDate)

		amtDaysAddedToRes := residentNow.AmtParkingDaysUsed - residentBefore.AmtParkingDaysUsed
		permitLength := int(lengthOfPermit.Hours() / 24)
		if !shouldAddDays && amtDaysAddedToRes != 0 {
			return fmt.Errorf("%s failed: added %d days when it shouldn't have. Permit length was: %d", testName, amtDaysAddedToRes, permitLength)
		} else if shouldAddDays && amtDaysAddedToRes != permitLength {
			return fmt.Errorf("%s failed: added incorrect amount of days: %d instead of %d", testName, amtDaysAddedToRes, permitLength)
		}

		return nil
	}

	for testName, test := range suite.testPermits {
		err := executeTest(testName, test)
		suite.NoError(err)
	}
}

func (suite permitRouterSuite) TestDelete_SubtractsResDays() {
	executeTest := func(testName string, newPermitReq newPermitReq) error {
		residentBefore, err := getTestResident(suite.testServer.URL, newPermitReq.ResidentId, suite.adminJWT)
		if err != nil {
			return fmt.Errorf("%s failed: %v", testName, err)
		}

		id, err := createTestPermit(suite.testServer.URL, suite.adminJWT, newPermitReq)
		if err != nil {
			return fmt.Errorf("%s failed: %v", testName, err)
		}

		err = deleteTestPermit(suite.testServer.URL, suite.adminJWT, id)
		if err != nil {
			return fmt.Errorf("%s failed: %v", testName, err)
		}

		residentNow, err := getTestResident(suite.testServer.URL, newPermitReq.ResidentId, suite.adminJWT)
		if err != nil {
			return fmt.Errorf("%s failed: %v", testName, err)
		}

		if residentBefore.AmtParkingDaysUsed != residentNow.AmtParkingDaysUsed {
			return fmt.Errorf("%s failed: did not subract days. Resident has %d instead of %d", testName, residentNow.AmtParkingDaysUsed, residentBefore.AmtParkingDaysUsed)
		}

		return nil
	}

	for testName, newPermitReq := range suite.testPermits {
		err := executeTest(testName, newPermitReq)
		suite.NoError(err)
	}
}

func (suite permitRouterSuite) TestDelete_AddsCarDays() {
}

func (suite permitRouterSuite) TestDelete_SubtractsCarDays() {
}

func (suite permitRouterSuite) TestCreate_AllFieldsMatch() {
}

func (suite permitRouterSuite) TestGetActivePermitsOfResident_Postive() {
	id, err := createTestPermit(suite.testServer.URL, suite.adminJWT, suite.testPermit)
	if err != nil {
		suite.NoError(err)
		return
	}
	defer deleteTestPermit(suite.testServer.URL, suite.adminJWT, id)

	endpoint := fmt.Sprintf("%s/api/resident/%s/permits/active", suite.testServer.URL, testResident.Id)
	responseBody, statusCode, err := authenticatedReq("GET", endpoint, nil, suite.adminJWT)
	if err != nil {
		suite.NoError(err)
		return
	}
	defer responseBody.Close()

	if statusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(responseBody)
		if err != nil {
			suite.NoError(fmt.Errorf("Error getting error response: %v", err))
			return
		}
		suite.NoError(fmt.Errorf("Bad response: %s", string(bodyBytes)))
		return
	}

	var permitsResponse listWithMetadata[models.Permit]
	if err := json.NewDecoder(responseBody).Decode(&permitsResponse); err != nil {
		suite.NoError(err)
		return
	} else if len(permitsResponse.Records) == 0 {
		suite.NotEmpty(permitsResponse.Records, "length of permits should not be zero")
		return
	}

	last := permitsResponse.Records[len(permitsResponse.Records)-1]
	suite.Equal(suite.testPermit.ResidentId, last.ResidentId)
	suite.Equal(suite.testPermit.Car.LicensePlate, last.Car.LicensePlate)
	suite.Empty(cmp.Diff(last.StartDate, suite.testPermit.StartDate))
	suite.Empty(cmp.Diff(last.EndDate, suite.testPermit.EndDate))
}

func (suite permitRouterSuite) TestGetMaxExceptions_Positive() {
	endpoint := fmt.Sprintf("%s/api/permits/exceptions?limit=%d", suite.testServer.URL, config.MaxLimit)
	responseBody, statusCode, err := authenticatedReq("GET", endpoint, nil, suite.adminJWT)
	if err != nil {
		suite.NoError(err)
		return
	}
	defer responseBody.Close()

	if statusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(responseBody)
		if err != nil {
			suite.NoError(fmt.Errorf("Error getting error response: %v", err))
			return
		}
		suite.NoError(fmt.Errorf("Bad response: %s", string(bodyBytes)))
		return
	}

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
func newTestPermit(residentId string, car newCarReq, exceptionReason string) newPermitReq {
	return newPermitReq{
		ResidentId:      residentId,
		Car:             car,
		StartDate:       time.Now().Truncate(time.Second),
		EndDate:         time.Now().Add(time.Duration(24) * time.Hour).Truncate(time.Second),
		ExceptionReason: exceptionReason,
	}
}

func createTestPermit(url string, adminJWT string, testPermit newPermitReq) (int, error) {
	requestBody, err := json.Marshal(testPermit)
	if err != nil {
		return 0, fmt.Errorf("Error marshalling: %v", err)
	}

	responseBody, statusCode, err := authenticatedReq("POST", url+"/api/permit", requestBody, adminJWT)
	if err != nil {
		return 0, fmt.Errorf("Error making request: %v", err)
	}
	defer responseBody.Close()

	if statusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(responseBody)
		if err != nil {
			return 0, fmt.Errorf("Error getting error response: %v", err)
		}
		return 0, fmt.Errorf("Bad response: %s", string(bodyBytes))
	}

	var newPermitResponse models.Permit
	if err := json.NewDecoder(responseBody).Decode(&newPermitResponse); err != nil {
		return 0, fmt.Errorf("Error decoding response: %v", err)
	}

	return newPermitResponse.Id, nil
}

func deleteTestPermit(url string, adminJWT string, id int) error {
	endpoint := fmt.Sprintf("%s/api/permit/%d", url, id)
	responseBody, statusCode, err := authenticatedReq("DELETE", endpoint, nil, adminJWT)
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

func getTestResident(url string, residentId string, adminJWT string) (models.Resident, error) {
	endpoint := fmt.Sprintf("%s/api/resident/%s", url, residentId)
	responseBody, statusCode, err := authenticatedReq("GET", endpoint, nil, adminJWT)
	if err != nil {
		return models.Resident{}, err
	}
	defer responseBody.Close()

	if statusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(responseBody)
		if err != nil {
			return models.Resident{}, fmt.Errorf("Error reading response after bad status code: %v", err)
		}

		return models.Resident{}, fmt.Errorf("Bad response getting resident: %s", string(bodyBytes))
	}

	var resident models.Resident
	if err := json.NewDecoder(responseBody).Decode(&resident); err != nil {
		return models.Resident{}, err
	}

	return resident, nil
}
