package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/app"
	"github.com/dannyvelas/lasvistas_api/errs"
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

func authenticatedReq[T, U any](method, url, accessToken string, requestBody *T) (parsedResp U, err error) {
	requestBytes := []byte{}
	if requestBody != nil {
		requestBytes, err = json.Marshal(requestBody)
		if err != nil {
			return parsedResp, fmt.Errorf("Error marshalling: %v", err)
		}
	}

	request, err := http.NewRequest(method, url, bytes.NewBuffer(requestBytes))
	if err != nil {
		return parsedResp, err
	}
	request.Header.Add("Authorization", "Bearer "+accessToken)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return parsedResp, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		defer response.Body.Close()
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return parsedResp, fmt.Errorf("Error reading response after non-200 status code, %d: %v", response.StatusCode, err)
		}

		return parsedResp, errs.NewApiErr(response.StatusCode, string(bodyBytes))
	}

	if err := json.NewDecoder(response.Body).Decode(&parsedResp); err != nil {
		return parsedResp, fmt.Errorf("Error decoding response: %v", err)
	}

	return parsedResp, nil
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
