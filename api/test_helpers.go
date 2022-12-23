package api

import (
	"bytes"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/app"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/util"
	"io"
	"net/http"
)

var (
	testResident = models.Resident{
		ID:        "B1234567",
		FirstName: "Daniel",
		LastName:  "Velasquez",
		Phone:     "1234567890",
		Email:     "email@example.com",
		Password:  "notapassword"}

	testResidentUnlimDays = models.Resident{
		ID:        "B7654321",
		FirstName: "Daniel",
		LastName:  "Velasquez",
		Phone:     "1234567890",
		Email:     "email2@example.com",
		Password:  "notapassword",
		UnlimDays: util.ToPtr(true)}
)

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

func createTestResidents(residentService app.ResidentService) error {
	if err := residentService.Create(testResident); err != nil {
		return err
	}

	if err := residentService.Create(testResidentUnlimDays); err != nil {
		return err
	}

	return nil
}

func deleteTestResidents(residentService app.ResidentService) error {
	if err := residentService.Delete(testResident.ID); err != nil {
		return err
	}

	if err := residentService.Delete(testResidentUnlimDays.ID); err != nil {
		return err
	}

	return nil
}
