package api

import (
	"bytes"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"io"
	"net/http"
	"net/http/httptest"
)

var (
	testResident = models.NewResident(
		"B1234567",
		"Daniel",
		"Velasquez",
		"1234567890",
		"email@example.com",
		"notapassword",
		false,
		0,
		0)
	testResidentUnlimDays = models.NewResident(
		"B7654321",
		"Daniel",
		"Velasquez",
		"1234567890",
		"email2@example.com",
		"notapassword",
		true,
		0,
		0)
)

func newTestServer(c config.Config, repos storage.Repos) *httptest.Server {
	router := NewRouter(c.Http(), c.Token(), c.OAuth(), config.DateFormat, repos)
	return httptest.NewServer(router)
}

func authenticatedReq(method string, url string, requestBytes []byte, accessToken string) (io.ReadCloser, int, error) {
	request, err := http.NewRequest(method, url, bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, 0, err
	}
	request.Header.Add("Authorization", "Bearer "+accessToken)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, 0, err
	}

	return response.Body, response.StatusCode, nil
}

func createTestResidents(testServerURL, accessToken string) error {
	createFn := func(testResident models.Resident) error {
		requestBody := []byte(fmt.Sprintf(`{
      "residentId": "%s",
      "firstName": "%s",
      "lastName": "%s",
      "phone": "%s",
      "email": "%s",
      "password":"%s",
      "unlimDays": %t
    }`,
			testResident.Id,
			testResident.FirstName,
			testResident.LastName,
			testResident.Phone,
			testResident.Email,
			testResident.Password,
			testResident.UnlimDays))

		responseBody, statusCode, err := authenticatedReq("POST", testServerURL+"/api/account", requestBody, accessToken)
		if err != nil {
			return fmt.Errorf("test_server.createTestResident: error sending request: %v", err)
		}
		defer responseBody.Close()

		if statusCode != http.StatusOK {
			bodyBytes, err := io.ReadAll(responseBody)
			if err != nil {
				return fmt.Errorf("test_server.createTestResident: error getting error response: %v", err)
			}
			return fmt.Errorf("test_server.createTestResident: Bad response: %s", string(bodyBytes))
		}

		return nil
	}

	for _, testResident := range []models.Resident{testResident, testResidentUnlimDays} {
		err := createFn(testResident)
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteTestResidents(testServerURL, accessToken string) error {
	deleteFn := func(testResident models.Resident) error {
		endpoint := fmt.Sprintf("%s/api/resident/%s", testServerURL, testResident.Id)
		responseBody, statusCode, err := authenticatedReq("DELETE", endpoint, nil, accessToken)
		if err != nil {
			return fmt.Errorf("test_server.deleteTestResident: error sending request: %v", err)
		}
		defer responseBody.Close()

		if statusCode != http.StatusOK {
			bodyBytes, err := io.ReadAll(responseBody)
			if err != nil {
				return fmt.Errorf("test_server.deleteTestResident: error getting error response: %v", err)
			}
			return fmt.Errorf("test_server.deleteTestResident: Bad response: %s", string(bodyBytes))
		}

		return nil
	}

	for _, testResident := range []models.Resident{testResident, testResidentUnlimDays} {
		err := deleteFn(testResident)
		if err != nil {
			return err
		}
	}

	return nil
}
