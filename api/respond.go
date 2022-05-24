package api

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

type emptyResponse struct {
	Ok bool `json:"ok"`
}

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

func respondInternalError(w http.ResponseWriter) {
	respondJSON(w, errInternalServerError.statusCode, errInternalServerError.message)
}

func respondError(w http.ResponseWriter, responseErr responseError) {
	respondJSON(w, responseErr.statusCode, responseErr.message)
}
