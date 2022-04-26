package api

import (
	"encoding/json"
	"errors"
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
		size := toUint(r.URL.Query().Get("size"))
		page := toUint(r.URL.Query().Get("page"))
		boundedSize, offset := getBoundedSizeAndOffset(size, page)

		activePermits, err := permitRepo.GetActive(boundedSize, offset)
		if err != nil {
			log.Error().Msgf("permit_router.GetActive: Error querying permitRepo: %v", err)
			respondError(w, errInternalServerError)
			return
		}

		respondJSON(w, http.StatusOK, activePermits)
	}
}

func getAll(permitRepo storage.PermitRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		size := toUint(r.URL.Query().Get("size"))
		page := toUint(r.URL.Query().Get("page"))
		boundedSize, offset := getBoundedSizeAndOffset(size, page)

		allPermits, err := permitRepo.GetAll(boundedSize, offset)
		if err != nil {
			log.Error().Msgf("permit_router.getAll: Error querying permitRepo: %v", err)
			respondError(w, errInternalServerError)
			return
		}

		respondJSON(w, http.StatusOK, allPermits)
	}
}

func create(permitRepo storage.PermitRepo, carRepo storage.CarRepo, dateFormat string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var createPermitReq createPermitReq
		if err := json.NewDecoder(r.Body).Decode(&createPermitReq); err != nil {
			respondError(w, errBadRequest)
			return
		}

		createPermit, err := createPermitReq.toModels()
		if err != nil {
			respondErrorWith(w, errBadRequest, err.Error())
			return
		}

		{
			// check if car exists
			existingCar, err := carRepo.GetByLicensePlate(createPermit.CreateCar.LicensePlate)
			if err != nil && !errors.Is(err, storage.ErrNoRows) { // unexpected error
				log.Error().Msgf("permit_router: Error querying carRepo: %v", err)
				respondError(w, errInternalServerError)
				return
			}

			// if car exists, check if it has active permits during dates requested
			if err == nil {
				activePermitsDuring, err := permitRepo.GetActiveOfCarDuring(
					existingCar.Id, createPermit.StartDate, createPermit.EndDate)
				if err != nil {
					log.Error().Msgf("permit_router.create: Error querying permitRepo: %v", err)
					respondError(w, errInternalServerError)
					return
				}

				if len(activePermitsDuring) != 0 {
					message := fmt.Sprintf("Cannot create a permit during dates %s and %s, "+
						"because this car has at least one active permit during that time.",
						createPermit.StartDate.Format(dateFormat),
						createPermit.EndDate.Format(dateFormat))
					respondErrorWith(w, errBadRequest, message)
					return
				}
			}
		}

		// TODO: check if resident exists

		respondJSON(w, 200, createPermit)
	}
}
