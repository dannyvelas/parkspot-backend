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
	testServer  *httptest.Server
	jwtToken    string
	residentId  string
	existingCar newCarReq
	newPermit   newPermitReq
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

	suite.residentId = "T1043321"
	suite.existingCar = newCarReq{"GBTYZME", "green", "ARCTIC CAT", "BEARCAT 2000 LT"}
	suite.newPermit = newPermitReq{
		ResidentId:      suite.residentId,
		Car:             suite.existingCar,
		StartDate:       time.Now().Truncate(time.Second),
		EndDate:         time.Now().Add(time.Duration(24) * time.Hour).Truncate(time.Second),
		ExceptionReason: ""}
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

func (suite permitRouterSuite) TestGetActivePermitsOfResident_Postive() {
	permitId, err := createTestPermit(suite.testServer.URL, suite.newPermit, suite.jwtToken)
	if err != nil {
		suite.NoError(err)
		return
	}
	defer deleteTestPermit(suite.testServer.URL, permitId, suite.jwtToken)

	endpoint := fmt.Sprintf("%s/api/resident/%s/permits/active", suite.testServer.URL, suite.residentId)

	request, err := http.NewRequest("GET", endpoint, nil)
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

	suite.Equal(http.StatusOK, response.StatusCode)

	var permitsResponse listWithMetadata[models.Permit]
	if err := json.NewDecoder(response.Body).Decode(&permitsResponse); err != nil {
		suite.NoError(err)
		return
	} else if len(permitsResponse.Records) == 0 {
		suite.NotEqual(len(permitsResponse.Records), 0, "no permits found")
		return
	}

	last := permitsResponse.Records[len(permitsResponse.Records)-1]
	suite.Equal(suite.newPermit.ResidentId, last.ResidentId)
	suite.Equal(suite.newPermit.Car.LicensePlate, last.Car.LicensePlate)
	suite.Empty(cmp.Diff(last.StartDate, suite.newPermit.StartDate))
	suite.Empty(cmp.Diff(last.EndDate, suite.newPermit.EndDate))
}

func createTestPermit(url string, body newPermitReq, jwtToken string) (int, error) {
	requestBody, err := json.Marshal(body)
	if err != nil {
		return 0, nil
	}

	request, err := http.NewRequest("POST", url+"/api/permit", bytes.NewBuffer(requestBody))
	if err != nil {
		return 0, nil
	}

	cookie := http.Cookie{Name: "jwt", Value: jwtToken, HttpOnly: true, Path: "/"}
	request.AddCookie(&cookie)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return 0, nil
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Status not OK")
	}

	var newPermitResponse models.Permit
	if err := json.NewDecoder(response.Body).Decode(&newPermitResponse); err != nil {
		return 0, nil
	}

	return newPermitResponse.Id, nil
}

func deleteTestPermit(url string, id int, jwtToken string) error {
	endpoint := fmt.Sprintf("%s/api/permit/%d", url, id)
	request, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return err
	}

	cookie := http.Cookie{Name: "jwt", Value: jwtToken, HttpOnly: true, Path: "/"}
	request.AddCookie(&cookie)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Status not OK")
	}

	return nil
}
