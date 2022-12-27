package api

import (
	"encoding/json"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

type message struct {
	Message string `json:"message"`
}

func respondJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Msgf("Error encoding response: %s", err)

		if _, err := io.WriteString(w, "Internal Server Error"); err != nil {
			log.Error().Msgf("Error sending Internal Server Error response: %q", err)
		}
	}
}

func respondError(w http.ResponseWriter, apiErr errs.ApiErr) {
	if apiErr.StatusCode == http.StatusInternalServerError {
		log.Error().Msg(apiErr.Error())
		respondJSON(w, apiErr.StatusCode, "Internal Server Error")
		return
	}

	respondJSON(w, apiErr.StatusCode, apiErr.Error())
}
