package api

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
)

func PermitsRouter(permitsRepo storage.PermitsRepo) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/active", getActive(permitsRepo))
		r.Get("/all", getAll(permitsRepo))
	}
}

func getActive(permitsRepo storage.PermitsRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Msg("Get Active Permits Endpoint")

		size := toUint(r.URL.Query().Get("size"))
		page := toUint(r.URL.Query().Get("page"))
		boundedSize, offset := getBoundedSizeAndOffset(size, page)

		activePermits, err := permitsRepo.GetActive(boundedSize, offset)
		if err != nil {
			err := fmt.Errorf("permits_router: GetActive: Error querying permitsRepo: %v", err)
			respondError(w, err, errInternalServerError)
			return
		}

		respondJSON(w, http.StatusOK, activePermits)
	}
}

func getAll(permitsRepo storage.PermitsRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Msg("Get All Endpoint")

		size := toUint(r.URL.Query().Get("size"))
		page := toUint(r.URL.Query().Get("page"))
		boundedSize, offset := getBoundedSizeAndOffset(size, page)

		allPermits, err := permitsRepo.GetAll(boundedSize, offset)
		if err != nil {
			err = fmt.Errorf("permits_router: getAll: Error querying permitsRepo: %v", err)
			respondError(w, err, errInternalServerError)
			return
		}

		respondJSON(w, http.StatusOK, allPermits)
	}
}

