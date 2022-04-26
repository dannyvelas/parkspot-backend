package api

import (
	"encoding/json"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
)

func PermitRouter(permitRepo storage.PermitRepo, carRepo storage.CarRepo, dateFormat string) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/active", getActive(permitRepo))
		r.Get("/all", getAll(permitRepo))
		r.Post("/create", create(permitRepo, carRepo, dateFormat))
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

func create(permitRepo storage.PermitRepo, carRepo storage.CarRepo, dateFormat string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var createPermitReq createPermitReq
		if err := json.NewDecoder(r.Body).Decode(&createPermitReq); err != nil {
			err = fmt.Errorf("permit_router.create: Error decoding credentials body: %v", err)
			respondError(w, err, errBadRequest)
			return
		}

		createPermit, err := createPermitReq.toModels()
		if err != nil {
			err := fmt.Errorf("permit_router.create: Invalid fields: %v", err)
			respondError(w, err, errBadRequest)
			return
		}

		// TODO: check if resident exists

		respondJSON(w, 200, createPermit)
	}
}
