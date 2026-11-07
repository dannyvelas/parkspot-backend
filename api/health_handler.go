package api

import (
	"net/http"

	"github.com/dannyvelas/parkspot-backend/storage"
)

type healthHandler struct {
	database storage.Database
}

func newHealthHandler(database storage.Database) healthHandler {
	return healthHandler{
		database: database,
	}
}

func (h healthHandler) up() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}
