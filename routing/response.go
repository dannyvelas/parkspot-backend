package routing

import (
	"encoding/json"
	"io"
	"net/http"
)

func RespondJson(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if data == nil {
		_, _ = w.Write([]byte(""))
		return
	}

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err = io.WriteString(w, `{"err": "Error parsing response"}`)
	}
}
