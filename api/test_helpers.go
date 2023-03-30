package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
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
	testCar = models.NewCar(
		"d1e0affb-14e7-4e9f-b8a3-70be7d49d063",
		testResident.ID,
		"lp1",
		"color",
		"make",
		"model",
		0)
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
