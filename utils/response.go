package utils

import (
	"encoding/json"
	"net/http"
)

func RespondJson(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	w.WriteHeader(statusCode)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(`{"err": "Error parsing response"}`))
		}
	} else {
		_, _ = w.Write([]byte(""))
	}
}
