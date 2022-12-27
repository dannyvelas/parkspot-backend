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

type permitHandler struct {
	permitService app.PermitService
}

func newPermitHandler(permitService app.PermitService) permitHandler {
	return permitHandler{
		permitService: permitService,
	}
}

func (h permitHandler) get(permitFilter models.PermitFilter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := util.ToPosInt(r.URL.Query().Get("limit"))
		page := util.ToPosInt(r.URL.Query().Get("page"))
		reversed := util.ToBool(r.URL.Query().Get("reversed"))
		search := r.URL.Query().Get("search")

		ctx := r.Context()
		accessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			respondError(w, *errs.Internalf("error getting access payload: %v", err))
			return
		}

		residentID := ""
		if accessPayload.Role == models.ResidentRole {
			residentID = accessPayload.ID
		}

		permitsWithMetadata, apiErr := h.permitService.GetAll(permitFilter, limit, page, reversed, search, residentID)
		if apiErr != nil {
			respondError(w, *apiErr)
			return
		}

		respondJSON(w, http.StatusOK, permitsWithMetadata)
	}
}

func (h permitHandler) getOne() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := util.ToPosInt(chi.URLParam(r, "id"))
		if id == 0 {
			respondError(w, *errs.BadRequest("id parameter cannot be empty"))
			return
		}

		permit, apiErr := h.permitService.GetOne(id)
		if apiErr != nil {
			respondError(w, *apiErr)
			return
		}

		respondJSON(w, http.StatusOK, permit)
	}
}

func (h permitHandler) create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newPermitReq models.Permit
		if err := json.NewDecoder(r.Body).Decode(&newPermitReq); err != nil {
			respondError(w, *errs.Malformed("NewPermitReq"))
			return
		}

		ctx := r.Context()
		accessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			respondError(w, *errs.Internalf("permit_router.createPermit: error getting access payload: %v", err))
			return
		}

		if accessPayload.Role == models.ResidentRole && newPermitReq.ExceptionReason != "" {
			message := "Residents cannot request parking permits with exceptions"
			respondError(w, *errs.BadRequest(message))
			return
		}

		createdPermit, apiErr := h.permitService.ValidateAndCreate(newPermitReq)
		if apiErr != nil {
			respondError(w, *apiErr)
			return
		}

		respondJSON(w, http.StatusOK, createdPermit)
	}
}

func (h permitHandler) deleteOne() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := util.ToPosInt(chi.URLParam(r, "id"))
		if id == 0 {
			respondError(w, *errs.BadRequest("id parameter cannot be empty"))
			return
		}

		apiErr := h.permitService.Delete(id)
		if apiErr != nil {
			respondError(w, *apiErr)
			return
		}

		respondJSON(w, http.StatusOK, message{"Successfully deleted permit"})
	}
}
