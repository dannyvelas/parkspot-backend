package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/dannyvelas/parkspot-backend/errs"
	"github.com/rs/zerolog/log"
)

type message struct {
	Message string `json:"message"`
}

func respondJSON(w http.ResponseWriter, r *http.Request, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		unhandledErrorsTotal.WithLabelValues(routePattern(r), errorLabel(err)).Inc()
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Msgf("Error encoding response: %s", err)

		if _, err := io.WriteString(w, "Internal Server Error"); err != nil {
			log.Error().Msgf("Error sending Internal Server Error response: %q", err)
		}
	}
}

func respondError(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr *errs.APIErr
	if !errors.As(err, &apiErr) {
		unhandledErrorsTotal.WithLabelValues(routePattern(r), errorLabel(err)).Inc()
		log.Error().Msg(err.Error())
		respondJSON(w, r, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	respondJSON(w, r, apiErr.StatusCode, apiErr.Error())
}

// routePattern returns chi's route pattern (e.g. "/api/permit/{id}"). Safe to fall back to the raw path here since this only ever runs for errors on routes that already matched.
func routePattern(r *http.Request) string {
	return routePatternOr(r, r.URL.Path)
}

func errorLabel(err error) string {
	var sentinel *errs.Sentinel
	if errors.As(err, &sentinel) {
		return sentinel.Label()
	}
	return "unknown"
}
