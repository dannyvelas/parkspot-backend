package api

import (
	"encoding/json"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/dannyvelas/lasvistas_api/util"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
)

func getOneCar(carRepo storage.CarRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !util.IsUUIDV4(id) {
			respondError(w, newErrBadRequest("id parameter is not a UUID"))
			return
		}

		car, err := carRepo.GetOne(id)
		if err != nil {
			log.Error().Msgf("car_router.getOne: Error getting car: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, car)
	}
}

func editCar(carRepo storage.CarRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !util.IsUUIDV4(id) {
			respondError(w, newErrBadRequest("id parameter is not a UUID"))
			return
		}

		var editCarReq editCarReq
		if err := json.NewDecoder(r.Body).Decode(&editCarReq); err != nil {
			respondError(w, newErrMalformed("EditCarReq"))
			return
		}

		if err := editCarReq.validate(); err != nil {
			respondError(w, newErrBadRequest(err.Error()))
			return
		}

		err := carRepo.Update(id, editCarReq.Color, editCarReq.Make, editCarReq.Model)
		if err != nil {
			log.Error().Msgf("car_router.editCar: Error updating car: %v", err)
			respondInternalError(w)
			return
		}

		car, err := carRepo.GetOne(id)
		if err != nil {
			log.Error().Msgf("car_router.editCar: Error getting car: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, car)
	}
}
