package api

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
		0)
	testResidentUnlimDays = models.NewResident(
		"B7654321",
		"Daniel",
		"Velasquez",
		"1234567890",
		"email2@example.com",
		"notapassword",
		false,
		0)
)

func newTestServer() (*httptest.Server, error) {
	config, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("Error loading config: %v", err.Error())
	}

	database, err := storage.NewDatabase(config.Postgres())
	if err != nil {
		return nil, fmt.Errorf("Failed to start database: %v", err)
	}

	// init repos
	adminRepo := storage.NewAdminRepo(database)
	permitRepo := storage.NewPermitRepo(database)
	carRepo := storage.NewCarRepo(database)
	residentRepo := storage.NewResidentRepo(database)
	visitorRepo := storage.NewVisitorRepo(database)

	// http setup
	httpConfig := config.Http()

	router := NewRouter(httpConfig, config.Token(), config.OAuth(), config.Constants().DateFormat(),
		adminRepo, permitRepo, carRepo, residentRepo, visitorRepo)

	testServer := httptest.NewServer(router)

	return testServer, nil
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

func getJWTToken(tokenConfig config.TokenConfig) (string, error) {
	jwtMiddleware := NewJWTMiddleware(tokenConfig)

	jwtToken, err := jwtMiddleware.newJWT("some-uuid", "Daniel", "Velasquez", "example@email.com", AdminRole)
	if err != nil {
		return "", fmt.Errorf("Failed to create JWT token: %v", err)
	}

	return jwtToken, nil
}

func createTestResidents(testServerURL, jwtToken string) error {
	createFn := func(testResident models.Resident) error {
		requestBody := []byte(fmt.Sprintf(`{
      "residentId": "%s",
      "firstName": "%s",
      "lastName": "%s",
      "phone": "%s",
      "email": "%s",
      "password":"%s"
    }`,
			testResident.Id,
			testResident.FirstName,
			testResident.LastName,
			testResident.Phone,
			testResident.Email,
			testResident.Password))

		responseBody, statusCode, err := authenticatedReq("POST", testServerURL+"/api/account", requestBody, jwtToken)
		if err != nil {
			return fmt.Errorf("test_server.createTestResident: error sending request: %v", err)
		}
		defer responseBody.Close()

		if statusCode != http.StatusOK {
			errorResponse, err := getErrorResponse(responseBody)
			if err != nil {
				return fmt.Errorf("test_server.createTestResident: error getting error response: %v", err)
			}
			return errors.New(errorResponse)
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

func deleteTestResidents(testServerURL, jwtToken string) error {
	deleteFn := func(testResident models.Resident) error {
		endpoint := fmt.Sprintf("%s/api/account/%s", testServerURL, testResident.Id)
		responseBody, statusCode, err := authenticatedReq("DELETE", endpoint, nil, jwtToken)
		if err != nil {
			return fmt.Errorf("test_server.deleteTestResident: error sending request: %v", err)
		}
		defer responseBody.Close()

		if statusCode != http.StatusOK {
			errorResponse, err := getErrorResponse(responseBody)
			if err != nil {
				return fmt.Errorf("test_server.deleteTestResident: error getting error response: %v", err)
			}

			return errors.New(errorResponse)
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

func getErrorResponse(responseBody io.ReadCloser) (string, error) {
	buf := new(strings.Builder)
	_, err := io.Copy(buf, responseBody)
	if err != nil {
		return "", fmt.Errorf("error decoding response error string: %v", err)
	}

	return buf.String(), nil
}
