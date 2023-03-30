package api

import (
	"encoding/json"
	"github.com/dannyvelas/lasvistas_api/app"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/util"
	"github.com/go-chi/chi/v5"
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
			respondError(w, errs.BadRequest("id parameter is not a UUID"))
			return
		}

		car, err := h.carService.GetOne(id)
		if err != nil {
			respondError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, car)
	}
}

func (h carHandler) edit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var editCarReq models.Car
		if err := json.NewDecoder(r.Body).Decode(&editCarReq); err != nil {
			respondError(w, errs.Malformed("EditCarReq"))
			return
		}

		car, err := h.carService.Update(editCarReq)
		if err != nil {
			respondError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, car)
	}
}

func (h carHandler) getOfResident() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cars, err := h.carService.GetOfResident(chi.URLParam(r, "id"))
		if err != nil {
			respondError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, cars)
	}
}
