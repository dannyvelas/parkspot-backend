package api

import (
	"encoding/json"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
)

func getOneCar(carRepo storage.CarRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !isUUIDV4(id) {
			respondErrorWith(w, errBadRequest, "id parameter is not a UUID")
			return
		}

		car, err := carRepo.GetOne(id)
		if err != nil {
			log.Error().Msgf("car_router.getOne: Error getting car: %v", err)
			respondError(w, errInternalServerError)
			return
		}

		respondJSON(w, http.StatusOK, car)
	}
}

func editCar(carRepo storage.CarRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !isUUIDV4(id) {
			respondErrorWith(w, errBadRequest, "id parameter is not a UUID")
			return
		}

		var editCarReq editCarReq
		if err := json.NewDecoder(r.Body).Decode(&editCarReq); err != nil {
			respondError(w, errBadRequest)
			return
		}

		if err := editCarReq.validate(); err != nil {
			respondErrorWith(w, errBadRequest, err.Error())
			return
		}

		editCarArgs := editCarReq.toEditCarArgs()

		err := carRepo.Update(id, editCarArgs)
		if err != nil {
			log.Error().Msgf("car_router.editCar: Error updating car: %v", err)
			respondError(w, errInternalServerError)
			return
		}

		car, err := carRepo.GetOne(id)
		if err != nil {
			log.Error().Msgf("car_router.editCar: Error getting car: %v", err)
			respondError(w, errInternalServerError)
			return
		}

		respondJSON(w, http.StatusOK, car)
	}
}
