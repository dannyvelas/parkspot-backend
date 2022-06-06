package api

import (
	"bytes"
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
	testServer             *httptest.Server
	jwtToken               string
	residentIdUnlimDays    string
	residentIdNonUnlimDays string
	existingCar            newCarReq
	customPermit           func(string, string) newPermitReq
	newPermit              newPermitReq
}

func TestPermitRouter(t *testing.T) {
	suite.Run(t, new(permitRouterSuite))
}

func (suite *permitRouterSuite) SetupSuite() {
	config := config.NewConfig()

	testServer, err := newTestServer()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	suite.testServer = testServer

	suite.jwtToken = func() string {
		jwtMiddleware := NewJWTMiddleware(config.Token())

		jwtToken, err := jwtMiddleware.newJWT("B1234567", "Daniel", "Velasquez", "example@email.com", AdminRole)
		if err != nil {
			log.Fatal().Msgf("Failed to create JWT token: %v", err)
		}

		return jwtToken
	}()

	suite.residentIdUnlimDays = "T1043321"
	suite.residentIdNonUnlimDays = "T2980699"
	suite.existingCar = newCarReq{"GBTYZME", "green", "ARCTIC CAT", "BEARCAT 2000 LT"}
	suite.customPermit = func(residentId, exceptionReason string) newPermitReq {
		return newPermitReq{
			ResidentId:      residentId,
			Car:             suite.existingCar,
			StartDate:       time.Now().Truncate(time.Second),
			EndDate:         time.Now().Add(time.Duration(24) * time.Hour).Truncate(time.Second),
			ExceptionReason: exceptionReason}
	}
	suite.newPermit = suite.customPermit(suite.residentIdUnlimDays, "")
}

func (suite permitRouterSuite) TearDownSuite() {
	suite.testServer.Close()
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
	type test struct {
		newPermitReq  newPermitReq
		shouldAddDays bool
	}
	tests := map[string]test{
		"NoUnlimDays,NoException": {
			newPermitReq:  suite.customPermit(suite.residentIdNonUnlimDays, ""),
			shouldAddDays: true,
		},
		"UnlimDays,NoException": {
			newPermitReq:  suite.customPermit(suite.residentIdUnlimDays, ""),
			shouldAddDays: false,
		},
		"NoUnlimDays,Exception": {
			newPermitReq:  suite.customPermit(suite.residentIdNonUnlimDays, "some reason"),
			shouldAddDays: false,
		},
		"UnlimDays,Exception": {
			newPermitReq:  suite.customPermit(suite.residentIdUnlimDays, "some reason"),
			shouldAddDays: false,
		},
	}

	executeTest := func(testName string, test test) error {
		residentBefore, err := getTestResident(suite.testServer.URL, test.newPermitReq.ResidentId, suite.jwtToken)
		if err != nil {
			return fmt.Errorf("%s failed: %v", testName, err)
		}

		permit, err := createTestPermit(suite.testServer.URL, test.newPermitReq, suite.jwtToken)
		if err != nil {
			return fmt.Errorf("%s failed: %v", testName, err)
		}
		defer deleteTestPermit(suite.testServer.URL, permit.Id, suite.jwtToken)

		residentNow, err := getTestResident(suite.testServer.URL, test.newPermitReq.ResidentId, suite.jwtToken)
		if err != nil {
			return fmt.Errorf("%s failed: %v", testName, err)
		}

		lengthOfPermit := test.newPermitReq.EndDate.Sub(test.newPermitReq.StartDate)

		amtDaysAddedToRes := residentNow.AmtParkingDaysUsed - residentBefore.AmtParkingDaysUsed
		permitLength := int(lengthOfPermit.Hours() / 24)
		if !test.shouldAddDays && amtDaysAddedToRes == permitLength {
			return fmt.Errorf("%s failed: added days when it shouldn't have", testName)
		} else if test.shouldAddDays && amtDaysAddedToRes != permitLength {
			return fmt.Errorf("%s failed: added incorrect amount of days: %d instead of %d", testName, amtDaysAddedToRes, permitLength)
		}

		return nil
	}

	for testName, test := range tests {
		err := executeTest(testName, test)
		suite.NoError(err)
	}
}

func (suite permitRouterSuite) TestDelete_SubtractsResDays() {
	newPermitReqs := map[string]newPermitReq{
		"NoUnlimDays,NoException": suite.customPermit(suite.residentIdNonUnlimDays, ""),
		"UnlimDays,NoException":   suite.customPermit(suite.residentIdUnlimDays, ""),
		"NoUnlimDays,Exception":   suite.customPermit(suite.residentIdNonUnlimDays, "some reason"),
		"UnlimDays,Exception":     suite.customPermit(suite.residentIdUnlimDays, "some reason"),
	}

	executeTest := func(testName string, newPermitReq newPermitReq) error {
		residentBefore, err := getTestResident(suite.testServer.URL, newPermitReq.ResidentId, suite.jwtToken)
		if err != nil {
			return fmt.Errorf("%s failed: %v", testName, err)
		}

		permit, err := createTestPermit(suite.testServer.URL, newPermitReq, suite.jwtToken)
		if err != nil {
			return fmt.Errorf("%s failed: %v", testName, err)
		}

		err = deleteTestPermit(suite.testServer.URL, permit.Id, suite.jwtToken)
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

	for testName, newPermitReq := range newPermitReqs {
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
	permit, err := createTestPermit(suite.testServer.URL, suite.newPermit, suite.jwtToken)
	if err != nil {
		suite.NoError(err)
		return
	}
	defer deleteTestPermit(suite.testServer.URL, permit.Id, suite.jwtToken)

	endpoint := fmt.Sprintf("%s/api/resident/%s/permits/active", suite.testServer.URL, suite.residentIdUnlimDays)
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

// helpers
func createTestPermit(url string, body newPermitReq, jwtToken string) (models.Permit, error) {
	requestBody, err := json.Marshal(body)
	if err != nil {
		return models.Permit{}, nil
	}

	responseBody, statusCode, err := authenticatedReq("POST", url+"/api/permit", requestBody, jwtToken)
	if err != nil {
		return models.Permit{}, nil
	}
	defer responseBody.Close()

	if statusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(responseBody)
		if err != nil {
			return models.Permit{}, fmt.Errorf("Error reading response after bad status code: %v", err)
		}

		return models.Permit{}, fmt.Errorf("Bad response getting resident: %s", string(bodyBytes))
	}

	var newPermitResponse models.Permit
	if err := json.NewDecoder(responseBody).Decode(&newPermitResponse); err != nil {
		return models.Permit{}, nil
	}

	return newPermitResponse, nil
}

func deleteTestPermit(url string, id int, jwtToken string) error {
	endpoint := fmt.Sprintf("%s/api/permit/%d", url, id)
	responseBody, statusCode, err := authenticatedReq("DELETE", endpoint, nil, jwtToken)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	if statusCode != http.StatusOK {
		return fmt.Errorf("Status not OK")
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

func authenticatedReq(method string, url string, requestBytes []byte, jwtToken string) (io.ReadCloser, int, error) {
	request, err := http.NewRequest(method, url, bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, 0, err
	}
	cookie := http.Cookie{Name: "jwt", Value: jwtToken, HttpOnly: true, Path: "/"}
	request.AddCookie(&cookie)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, 0, err
	}

	return response.Body, response.StatusCode, nil
}
