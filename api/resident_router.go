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

type ResidentHandler struct {
	residentService app.ResidentService
}

func NewResidentHandler(residentService app.ResidentService) ResidentHandler {
	return ResidentHandler{
		residentService: residentService,
	}
}

func (h ResidentHandler) GetAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := util.ToPosInt(r.URL.Query().Get("limit"))
		page := util.ToPosInt(r.URL.Query().Get("page"))
		search := r.URL.Query().Get("search")

		residentsWithMetadata, err := h.residentService.GetAll(limit, page, search)
		if err != nil {
			log.Error().Msgf("error getting residents with metadata: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, residentsWithMetadata)
	}
}

func (h ResidentHandler) GetOne() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			respondError(w, newErrBadRequest("id parameter cannot be empty"))
			return
		}

		resident, err := h.residentService.GetOne(id)
		if err != nil && !errors.Is(err, app.ErrNotFound) {
			log.Error().Msgf("Error getting resident: %v", err)
			respondInternalError(w)
			return
		} else if errors.Is(err, app.ErrNotFound) {
			respondError(w, newErrNotFound("resident"))
			return
		}

		respondJSON(w, http.StatusOK, resident)
	}
}

func (h ResidentHandler) Edit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := models.IsResidentID(id); err != nil {
			respondError(w, newErrBadRequest(err.Error()))
			return
		}

		var editResidentReq editResidentReq
		if err := json.NewDecoder(r.Body).Decode(&editResidentReq); err != nil {
			respondError(w, newErrMalformed("EditResidentReq"))
			return
		}

		if err := editResidentReq.validate(); err != nil {
			respondError(w, newErrBadRequest(err.Error()))
			return
		}

		desiredResident := models.EditResident{
			FirstName:          editResidentReq.FirstName,
			LastName:           editResidentReq.LastName,
			Phone:              editResidentReq.Phone,
			Email:              editResidentReq.Email,
			Password:           editResidentReq.Password,
			UnlimDays:          editResidentReq.UnlimDays,
			AmtParkingDaysUsed: editResidentReq.AmtParkingDaysUsed,
		}

		resident, err := h.residentService.Update(id, desiredResident)
		if err != nil {
			log.Error().Msgf("Error updating resident: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, resident)
	}
}

func (h ResidentHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		err := h.residentService.Delete(id)
		if errors.Is(err, app.ErrNotFound) {
			respondError(w, newErrNotFound("resident"))
			return
		} else if err != nil {
			log.Error().Msgf("error deleting resident with resident service: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, message{"Successfully deleted resident"})
	}
}

func (h ResidentHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload newResidentReq
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			respondError(w, newErrMalformed("NewResidentReq"))
			return
		}

		if err := payload.validate(); err != nil {
			respondError(w, newErrBadRequest(err.Error()))
			return
		}

		desiredRes := models.CreateResident(payload)
		err := h.residentService.Create(desiredRes)
		if errors.Is(err, app.ErrAlreadyExists) {
			respondError(w, newErrBadRequest(err.Error()))
			return
		} else if err != nil {
			log.Error().Msgf("error getting resident by id: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, message{"Resident successfully created."})
	}
}
