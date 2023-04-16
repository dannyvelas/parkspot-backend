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
			respondError(w, err)
			return
		}
		residents := util.MapSlice(residentsWithMetadata.Records, h.removeHash)
		residentsWithMetadata.Records = residents

		respondJSON(w, http.StatusOK, residentsWithMetadata)
	}
}

func (h residentHandler) getOne() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resident, err := h.residentService.GetOne(chi.URLParam(r, "id"))
		if err != nil {
			respondError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, h.removeHash(resident))
	}
}

func (h residentHandler) edit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var editResidentReq models.Resident
		if err := json.NewDecoder(r.Body).Decode(&editResidentReq); err != nil {
			respondError(w, errs.Malformed("EditResidentReq"))
			return
		}

		if editResidentReq.Password != "" {
			respondError(w, errs.BadRequest("Resident passwords cannot be edited. These can only be changed if a resident requests a password reset."))
			return
		}

		resident, err := h.residentService.Update(editResidentReq)
		if err != nil {
			respondError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, h.removeHash(resident))
	}
}

func (h residentHandler) deleteOne() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h.residentService.Delete(chi.URLParam(r, "id")); err != nil {
			respondError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, message{"Successfully deleted resident"})
	}
}

func (h residentHandler) create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload models.Resident
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			respondError(w, errs.Malformed("NewResidentReq"))
			return
		}

		createdRes, err := h.residentService.Create(payload)
		if err != nil {
			respondError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, h.removeHash(createdRes))
	}
}

func (h residentHandler) removeHash(resident models.Resident) models.Resident {
	resident.Password = ""
	return resident
}
