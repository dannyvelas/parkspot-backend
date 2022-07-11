package api

import (
	"encoding/json"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
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
	testServer  *httptest.Server
	jwtToken    string
	newCar      newCarReq
	testPermits map[string]newPermitReq
	newPermit   newPermitReq
}

func TestPermitRouter(t *testing.T) {
	suite.Run(t, new(permitRouterSuite))
}

func (suite *permitRouterSuite) SetupSuite() {
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

	suite.newCar = newCarReq{"ABC123", "red", "toyota", "tercel"}

	suite.testPermits = map[string]newPermitReq{
		"NoUnlimDays,NoException": newTestPermit(testResident.Id, suite.newCar, ""),
		"UnlimDays,NoException":   newTestPermit(testResidentUnlimDays.Id, suite.newCar, ""),
		"NoUnlimDays,Exception":   newTestPermit(testResident.Id, suite.newCar, "some exception reason"),
		"UnlimDays,Exception":     newTestPermit(testResidentUnlimDays.Id, suite.newCar, "some exception reason"),
	}

	suite.newPermit = newTestPermit(testResidentUnlimDays.Id, suite.newCar, "")
}

func (suite permitRouterSuite) TearDownSuite() {
	defer suite.testServer.Close()

	err := deleteTestResidents(suite.testServer.URL, suite.jwtToken)
	if err != nil {
		log.Error().Msg("auth_router_test.TearDownSuite: " + err.Error())
		return
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
	responseBody, statusCode, err := authenticatedReq("POST", suite.testServer.URL+"/api/permit", requestBody, suite.jwtToken)
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
	responseBody, statusCode, err := authenticatedReq("POST", suite.testServer.URL+"/api/permit", requestBody, suite.jwtToken)
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
		residentBefore, err := getTestResident(suite.testServer.URL, testPermit.ResidentId, suite.jwtToken)
		if err != nil {
			return fmt.Errorf("%s failed: %v", testName, err)
		}

		id, err := createTestPermit(suite.testServer.URL, suite.jwtToken, testPermit)
		if err != nil {
			return fmt.Errorf("%s failed: %v", testName, err)
		}
		defer deleteTestPermit(suite.testServer.URL, suite.jwtToken, id)

		residentNow, err := getTestResident(suite.testServer.URL, testPermit.ResidentId, suite.jwtToken)
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
		residentBefore, err := getTestResident(suite.testServer.URL, newPermitReq.ResidentId, suite.jwtToken)
		if err != nil {
			return fmt.Errorf("%s failed: %v", testName, err)
		}

		id, err := createTestPermit(suite.testServer.URL, suite.jwtToken, newPermitReq)
		if err != nil {
			return fmt.Errorf("%s failed: %v", testName, err)
		}

		err = deleteTestPermit(suite.testServer.URL, suite.jwtToken, id)
		if err != nil {
			return fmt.Errorf("%s failed: %v", testName, err)
		}

		residentNow, err := getTestResident(suite.testServer.URL, newPermitReq.ResidentId, suite.jwtToken)
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
	id, err := createTestPermit(suite.testServer.URL, suite.jwtToken, suite.newPermit)
	if err != nil {
		suite.NoError(err)
		return
	}
	defer deleteTestPermit(suite.testServer.URL, suite.jwtToken, id)

	endpoint := fmt.Sprintf("%s/api/resident/%s/permits/active", suite.testServer.URL, testResidentUnlimDays.Id)
	responseBody, statusCode, err := authenticatedReq("GET", endpoint, nil, suite.jwtToken)
	if err != nil {
		suite.NoError(err)
		return
	}
	defer responseBody.Close()

	suite.Equal(http.StatusOK, statusCode)

	var permitsResponse listWithMetadata[models.Permit]
	if err := json.NewDecoder(responseBody).Decode(&permitsResponse); err != nil {
		suite.NoError(err)
		return
	} else if len(permitsResponse.Records) == 0 {
		suite.NotEmpty(permitsResponse.Records, "length of permits should not be zero")
		return
	}

	last := permitsResponse.Records[len(permitsResponse.Records)-1]
	suite.Equal(suite.newPermit.ResidentId, last.ResidentId)
	suite.Equal(suite.newPermit.Car.LicensePlate, last.Car.LicensePlate)
	suite.Empty(cmp.Diff(last.StartDate, suite.newPermit.StartDate))
	suite.Empty(cmp.Diff(last.EndDate, suite.newPermit.EndDate))
}

func (suite permitRouterSuite) TestGetMaxExceptions_Positive() {
	endpoint := fmt.Sprintf("%s/api/permits/exceptions?limit=%d", suite.testServer.URL, config.MaxLimit)
	responseBody, statusCode, err := authenticatedReq("GET", endpoint, nil, suite.jwtToken)
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

func createTestPermit(url string, jwtToken string, testPermit newPermitReq) (int, error) {
	requestBody, err := json.Marshal(testPermit)
	if err != nil {
		return 0, fmt.Errorf("Error marshalling: %v", err)
	}

	responseBody, statusCode, err := authenticatedReq("POST", url+"/api/permit", requestBody, jwtToken)
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

func deleteTestPermit(url string, jwtToken string, id int) error {
	endpoint := fmt.Sprintf("%s/api/permit/%d", url, id)
	responseBody, statusCode, err := authenticatedReq("DELETE", endpoint, nil, jwtToken)
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

func getTestResident(url string, residentId string, jwtToken string) (models.Resident, error) {
	endpoint := fmt.Sprintf("%s/api/resident/%s", url, residentId)
	responseBody, statusCode, err := authenticatedReq("GET", endpoint, nil, jwtToken)
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
