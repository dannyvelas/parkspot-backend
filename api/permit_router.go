package api

import (
	"encoding/json"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
)

func PermitRouter(permitRepo storage.PermitRepo, carRepo storage.CarRepo) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/active", getActive(permitRepo))
		r.Get("/all", getAll(permitRepo))
		r.Get("/create", create(permitRepo, carRepo))
	}
}

func getActive(permitRepo storage.PermitRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Msg("Get Active Permit Endpoint")

		size := toUint(r.URL.Query().Get("size"))
		page := toUint(r.URL.Query().Get("page"))
		boundedSize, offset := getBoundedSizeAndOffset(size, page)

		activePermits, err := permitRepo.GetActive(boundedSize, offset)
		if err != nil {
			err := fmt.Errorf("permit_router.GetActive: Error querying permitRepo: %v", err)
			respondError(w, err, errInternalServerError)
			return
		}

		respondJSON(w, http.StatusOK, activePermits)
	}
}

func getAll(permitRepo storage.PermitRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Msg("Get All Endpoint")

		size := toUint(r.URL.Query().Get("size"))
		page := toUint(r.URL.Query().Get("page"))
		boundedSize, offset := getBoundedSizeAndOffset(size, page)

		allPermits, err := permitRepo.GetAll(boundedSize, offset)
		if err != nil {
			err := fmt.Errorf("permit_router.getAll: Error querying permitRepo: %v", err)
			respondError(w, err, errInternalServerError)
			return
		}

		respondJSON(w, http.StatusOK, allPermits)
	}
}

func create(permitRepo storage.PermitRepo, carRepo storage.CarRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var permit models.Permit
		if err := json.NewDecoder(r.Body).Decode(&permit); err != nil {
			err = fmt.Errorf("permit_router.create: Error decoding credentials body: %v", err)
			respondError(w, err, errBadRequest)
			return
		}

		if err := permit.Validate(); err != nil {
			err := fmt.Errorf("permit_router.create: Invalid fields: %v", err)
			respondError(w, err, errBadRequest)
			return
		}

		// TODO: check if resident exists

		// TODO: check if licensePlate has active permit
	}
}
