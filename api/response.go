package api

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

func respondJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Msgf("Error encoding response: %s", err)

		if _, err := io.WriteString(w, errInternalServerError.Error()); err != nil {
			log.Error().Msgf("Error sending Internal Server Error response: %q", err)
		}
	}
}

func respondError(w http.ResponseWriter, responseErr responseError) {
	respondJSON(w, responseErr.statusCode, responseErr.message)
}

func respondErrorWith(w http.ResponseWriter, responseErr responseError, message string) {
	newResponseErr := responseErr
	responseErr.message = responseErr.message + ". " + message

	respondError(w, newResponseErr)
}
