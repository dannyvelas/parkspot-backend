package api

import (
	"encoding/json"
	"fmt"
	"github.com/dannyvelas/parkspot-backend/app"
	"github.com/dannyvelas/parkspot-backend/errs"
	"github.com/dannyvelas/parkspot-backend/models"
	"github.com/dannyvelas/parkspot-backend/util"
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

func (h carHandler) get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := util.ToPosInt(r.URL.Query().Get("limit"))
		page := util.ToPosInt(r.URL.Query().Get("page"))
		reversed := util.ToBool(r.URL.Query().Get("reversed"))
		search := r.URL.Query().Get("search")

		ctx := r.Context()
		accessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			respondError(w, fmt.Errorf("error getting access payload: %v", err))
			return
		}

		residentID := ""
		if accessPayload.Role == models.ResidentRole {
			residentID = accessPayload.ID
		}

		carsWithMetadata, err := h.carService.GetAll(limit, page, reversed, search, residentID)
		if err != nil {
			respondError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, carsWithMetadata)
	}
}

func (h carHandler) deleteOne() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !util.IsUUIDV4(id) {
			respondError(w, errs.IDNotUUID)
			return
		}

		ctx := r.Context()
		accessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			respondError(w, fmt.Errorf("car_handler.deleteOne: error getting access payload: %v", err))
			return
		}

		carToDelete, err := h.carService.GetOne(id)
		if err != nil {
			respondError(w, err)
			return
		}

		if accessPayload.Role == models.ResidentRole && carToDelete.ResidentID != accessPayload.ID {
			respondError(w, errs.NewUnauthorized("resident cannot delete car of another resident"))
			return
		}

		if err := h.carService.Delete(id); err != nil {
			respondError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, message{"Successfully deleted car"})
	}
}

func (h carHandler) getOne() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		car, err := h.carService.GetOne(chi.URLParam(r, "id"))
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
		if !util.IsUUIDV4(editCarReq.ID) {
			respondError(w, errs.IDNotUUID)
			return
		}

		ctx := r.Context()
		accessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			respondError(w, fmt.Errorf("car_handler.edit(): error getting access payload: %v", err))
			return
		}

		carToEdit, err := h.carService.GetOne(editCarReq.ID)
		if err != nil {
			respondError(w, err)
			return
		}

		if accessPayload.Role == models.ResidentRole && carToEdit.ResidentID != accessPayload.ID {
			respondError(w, errs.NewUnauthorized("resident cannot edit car of another resident"))
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

func (h carHandler) create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var desiredCar models.Car
		if err := json.NewDecoder(r.Body).Decode(&desiredCar); err != nil {
			respondError(w, errs.Malformed("New Car Request"))
			return
		}

		ctx := r.Context()
		accessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			respondError(w, fmt.Errorf("car_handler.create(): error getting access payload: %v", err))
			return
		}

		if accessPayload.Role == models.ResidentRole {
			if desiredCar.ResidentID != "" && desiredCar.ResidentID != accessPayload.ID {
				respondError(w, errs.NewUnauthorized("resident cannot create car for another resident"))
				return
			}
			if desiredCar.ResidentID == "" {
				desiredCar.ResidentID = accessPayload.ID
			}
		}

		car, err := h.carService.Create(desiredCar)
		if err != nil {
			respondError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, car)
	}
}

func (h carHandler) getOfResident() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		residentID := chi.URLParam(r, "id")
		if residentID == "" {
			respondError(w, errs.MissingIDField)
			return
		}

		cars, err := h.carService.GetAll(0, 0, false, "", residentID)
		if err != nil {
			respondError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, cars)
	}
}
