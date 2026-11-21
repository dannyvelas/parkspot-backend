package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/dannyvelas/parkspot-backend/errs"
	"github.com/go-chi/chi/v5"
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
		// "unknown": T004 will replace this with an errors.Is chain against
		// every known sentinel; until then this is CounterVec's only valid
		// 2-value call, not a real classification.
		unhandledErrorsTotal.WithLabelValues(routePattern(r), "unknown").Inc()
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
		// "unknown": same placeholder as respondJSON's encode-failure branch
		// above, until T004 lands.
		unhandledErrorsTotal.WithLabelValues(routePattern(r), "unknown").Inc()
		log.Error().Msg(err.Error())
		respondJSON(w, r, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	respondJSON(w, r, apiErr.StatusCode, apiErr.Error())
}

// routePattern returns chi's route pattern (e.g. "/api/permit/{id}"),
func routePattern(r *http.Request) string {
	rctx := chi.RouteContext(r.Context())
	if rctx == nil {
		return r.URL.Path
	}

	pattern := rctx.RoutePattern()
	if pattern == "" {
		return r.URL.Path
	}

	return pattern
}
