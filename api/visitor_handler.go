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

type visitorHandler struct {
	visitorService app.VisitorService
}

func newVisitorHandler(visitorService app.VisitorService) visitorHandler {
	return visitorHandler{
		visitorService: visitorService,
	}
}

func (h visitorHandler) getActive() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := util.ToPosInt(r.URL.Query().Get("limit"))
		page := util.ToPosInt(r.URL.Query().Get("page"))
		search := r.URL.Query().Get("search")

		ctx := r.Context()
		accessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			respondError(w, *errs.Internalf("visitor_router.getVisitorsOfResident: %v", err))
			return
		}

		residentID := ""
		if accessPayload.Role == models.ResidentRole {
			residentID = accessPayload.ID
		}

		visitorsWithMetadata, apiErr := h.visitorService.GetActive(limit, page, search, residentID)
		if apiErr != nil {
			respondError(w, *apiErr)
			return
		}

		respondJSON(w, http.StatusOK, visitorsWithMetadata)
	}
}

func (h visitorHandler) create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		accessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			respondError(w, *errs.Internalf("error getting access payload in visitor handler: %v", err))
			return
		}

		var payload models.Visitor
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			respondError(w, *errs.Malformed("NewVisitorReq"))
			return
		}

		visitor, apiErr := h.visitorService.Create(accessPayload.ID, payload)
		if apiErr != nil {
			respondError(w, *apiErr)
			return
		}

		respondJSON(w, http.StatusOK, visitor)
	}
}

func (h visitorHandler) deleteOne() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !util.IsUUIDV4(id) {
			respondError(w, *errs.BadRequest("id parameter is not a UUID"))
			return
		}

		if apiErr := h.visitorService.Delete(id); apiErr != nil {
			respondError(w, *apiErr)
			return
		}

		respondJSON(w, http.StatusOK, message{"Successfully deleted visitor"})
	}
}
