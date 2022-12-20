package api

import (
	"encoding/json"
	"errors"
	"github.com/dannyvelas/lasvistas_api/app"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/util"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
)

type carHandler struct {
	carService app.CarService
}

func newCarHandler(carService app.CarService) carHandler {
	return carHandler{
		carService: carService,
	}
}

func (h carHandler) getOne() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !util.IsUUIDV4(id) {
			respondError(w, newErrBadRequest("id parameter is not a UUID"))
			return
		}

		car, err := h.carService.GetOne(id)
		if err != nil && !errors.Is(err, app.ErrNotFound) {
			log.Error().Msgf("Error getting one car from carService: %v", err)
			respondInternalError(w)
			return
		} else if errors.Is(err, app.ErrNotFound) {
			respondError(w, newErrNotFound("carr"))
			return
		}

		respondJSON(w, http.StatusOK, car)
	}
}

func (h carHandler) edit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !util.IsUUIDV4(id) {
			respondError(w, newErrBadRequest("id parameter is not a UUID"))
			return
		}

		var editCarReq models.Car
		if err := json.NewDecoder(r.Body).Decode(&editCarReq); err != nil {
			respondError(w, newErrMalformed("EditCarReq"))
			return
		}

		if err := editCarReq.ValidateEdit(); err != nil {
			respondError(w, newErrBadRequest(err.Error()))
			return
		}

		car, err := h.carService.Update(id, editCarReq)
		if err != nil {
			log.Error().Msgf("error updating car from carService: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, car)
	}
}
