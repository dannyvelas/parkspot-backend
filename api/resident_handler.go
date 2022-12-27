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

		residentsWithMetadata, apiErr := h.residentService.GetAll(limit, page, search)
		if apiErr != nil {
			respondError(w, *apiErr)
			return
		}

		respondJSON(w, http.StatusOK, residentsWithMetadata)
	}
}

func (h residentHandler) getOne() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			respondError(w, *errs.BadRequest("id parameter cannot be empty"))
			return
		}

		resident, apiErr := h.residentService.GetOne(id)
		if apiErr != nil {
			respondError(w, *apiErr)
			return
		}

		respondJSON(w, http.StatusOK, resident)
	}
}

func (h residentHandler) edit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := models.IsResidentID(id); err != nil {
			respondError(w, *errs.BadRequest(err.Error()))
			return
		}

		var editResidentReq models.Resident
		if err := json.NewDecoder(r.Body).Decode(&editResidentReq); err != nil {
			respondError(w, *errs.Malformed("EditResidentReq"))
			return
		}

		resident, apiErr := h.residentService.Update(id, editResidentReq)
		if apiErr != nil {
			respondError(w, *apiErr)
			return
		}

		respondJSON(w, http.StatusOK, resident)
	}
}

func (h residentHandler) deleteOne() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			respondError(w, *errs.BadRequest("id parameter cannot be empty"))
			return
		}

		if apiErr := h.residentService.Delete(id); apiErr != nil {
			respondError(w, *apiErr)
			return
		}

		respondJSON(w, http.StatusOK, message{"Successfully deleted resident"})
	}
}

func (h residentHandler) create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload models.Resident
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			respondError(w, *errs.Malformed("NewResidentReq"))
			return
		}

		if apiErr := h.residentService.Create(payload); apiErr != nil {
			respondError(w, *apiErr)
			return
		}

		respondJSON(w, http.StatusOK, message{"Resident successfully created."})
	}
}
