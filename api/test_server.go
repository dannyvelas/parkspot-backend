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

func authenticatedReq(method string, url string, requestBytes []byte, accessToken string) (io.ReadCloser, error) {
	request, err := http.NewRequest(method, url, bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", "Bearer "+accessToken)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		defer response.Body.Close()
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("Error reading response after non-200 status code, %d: %v", response.StatusCode, err)
		}

		return nil, responseError{response.StatusCode, string(bodyBytes)}
	}

	return response.Body, nil
}

// Create Resident Funcs
func createTestResidents(testServerURL, accessToken string) error {
	err := hitCreateResidentEndpoint(testServerURL, accessToken, testResident)
	if err != nil {
		return err
	}

	err = hitCreateResidentEndpoint(testServerURL, accessToken, testResidentUnlimDays)
	if err != nil {
		return err
	}

	return nil
}

func hitCreateResidentEndpoint(testServerURL, accessToken string, resident models.Resident) error {
	requestBody := []byte(fmt.Sprintf(`{
      "residentId": "%s",
      "firstName": "%s",
      "lastName": "%s",
      "phone": "%s",
      "email": "%s",
      "password":"%s",
      "unlimDays": %t
    }`,
		resident.Id,
		resident.FirstName,
		resident.LastName,
		resident.Phone,
		resident.Email,
		resident.Password,
		resident.UnlimDays))

	responseBody, err := authenticatedReq("POST", testServerURL+"/api/account", requestBody, accessToken)
	if err != nil {
		return fmt.Errorf("test_server.hitCreateResidentEndpoint: request error: %v", err)
	}
	defer responseBody.Close()

	return nil
}

// Delete Resident Funcs
func deleteTestResidents(testServerURL, accessToken string) error {
	err := hitDeleteResidentEndpoint(testServerURL, accessToken, testResident.Id)
	if err != nil {
		return err
	}

	err = hitDeleteResidentEndpoint(testServerURL, accessToken, testResidentUnlimDays.Id)
	if err != nil {
		return err
	}

	return nil
}

func hitDeleteResidentEndpoint(testServerURL, accessToken string, residentID string) error {
	endpoint := fmt.Sprintf("%s/api/resident/%s", testServerURL, residentID)
	responseBody, err := authenticatedReq("DELETE", endpoint, nil, accessToken)
	if err != nil {
		return fmt.Errorf("test_server.deleteTestResident: req/res error: %v", err)
	}
	defer responseBody.Close()

	return nil
}
