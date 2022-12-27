package api

import (
	"encoding/json"
	"errors"
	"github.com/dannyvelas/lasvistas_api/app"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/util"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
)

type residentHandler struct {
	residentService app.ResidentService
}

func newResidentHandler(residentService app.ResidentService) residentHandler {
	return residentHandler{
		residentService: residentService,
	}
}

func (h residentHandler) getAll() http.HandlerFunc {
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

func (h residentHandler) getOne() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			respondError(w, newErrBadRequest("id parameter cannot be empty"))
			return
		}

		resident, err := h.residentService.GetOne(id)
		var apiErr errs.ApiErr
		if errors.Is(err, errs.NotFound) {
			respondError(w, newErrNotFound("resident"))
			return
		} else if errors.As(err, &apiErr) {
			respondError(w, newErrBadRequest(err.Error()))
			return
		} else if err != nil {
			log.Error().Msgf("Error getting resident: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, resident)
	}
}

func (h residentHandler) edit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := models.IsResidentID(id); err != nil {
			respondError(w, newErrBadRequest(err.Error()))
			return
		}

		var editResidentReq models.Resident
		if err := json.NewDecoder(r.Body).Decode(&editResidentReq); err != nil {
			respondError(w, newErrMalformed("EditResidentReq"))
			return
		}

		resident, err := h.residentService.Update(id, editResidentReq)
		var apiErr errs.ApiErr
		if errors.As(err, &apiErr) {
			respondError(w, newErrBadRequest(err.Error()))
			return
		} else if err != nil {
			log.Error().Msgf("Error updating resident: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, resident)
	}
}

func (h residentHandler) deleteOne() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		err := h.residentService.Delete(id)
		var apiErr errs.ApiErr
		if errors.Is(err, errs.NotFound) {
			respondError(w, newErrNotFound("resident"))
			return
		} else if errors.As(err, &apiErr) {
			respondError(w, newErrBadRequest(err.Error()))
			return
		} else if err != nil {
			log.Error().Msgf("Error getting resident: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, message{"Successfully deleted resident"})
	}
}

func (h residentHandler) create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload models.Resident
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			respondError(w, newErrMalformed("NewResidentReq"))
			return
		}

		err := h.residentService.Create(payload)
		var apiErr errs.ApiErr
		if errors.As(err, &apiErr) {
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
