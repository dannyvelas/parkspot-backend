package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/errs"
	"io"
	"net/http"
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
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return parsedResp, fmt.Errorf("Error reading response after non-200 status code, %d: %v", response.StatusCode, err)
		}

		if response.StatusCode == 405 {
			return parsedResp, errs.NewApiErr(response.StatusCode, "Method Not Allowed")
		}

		return parsedResp, errs.NewApiErr(response.StatusCode, string(bodyBytes))
	}

	if err := json.NewDecoder(response.Body).Decode(&parsedResp); err != nil {
		return parsedResp, fmt.Errorf("Error decoding response: %v", err)
	}

	return parsedResp, nil
}
