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

func respondError(w http.ResponseWriter, internalErr error, apiErr apiError) {
	statusCode, message := apiErr.apiError()
	if statusCode == http.StatusInternalServerError {
		log.Error().Msg(internalErr.Error())
	} else {
		log.Debug().Msg(internalErr.Error())
	}
	respondJSON(w, statusCode, message)
}
