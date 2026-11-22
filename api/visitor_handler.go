package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dannyvelas/parkspot-backend/app"
	"github.com/dannyvelas/parkspot-backend/errs"
	"github.com/dannyvelas/parkspot-backend/models"
	"github.com/dannyvelas/parkspot-backend/util"
	"github.com/go-chi/chi/v5"
)

type visitorHandler struct {
	visitorService app.VisitorService
}

func newVisitorHandler(visitorService app.VisitorService) visitorHandler {
	return visitorHandler{
		visitorService: visitorService,
	}
}

func (h visitorHandler) get(status models.Status) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := util.ToPosInt(r.URL.Query().Get("limit"))
		page := util.ToPosInt(r.URL.Query().Get("page"))
		search := r.URL.Query().Get("search")

		ctx := r.Context()
		accessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			respondError(w, r, fmt.Errorf("visitor_router.getVisitorsOfResident: %w", err))
			return
		}

		residentID := ""
		if accessPayload.Role == models.ResidentRole {
			residentID = accessPayload.ID
		}

		visitorsWithMetadata, err := h.visitorService.Get(status, limit, page, search, residentID)
		if err != nil {
			respondError(w, r, err)
			return
		}

		respondJSON(w, r, http.StatusOK, visitorsWithMetadata)
	}
}

func (h visitorHandler) create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var desiredVisitor models.Visitor
		if err := json.NewDecoder(r.Body).Decode(&desiredVisitor); err != nil {
			respondError(w, r, Malformed("New Visitor Request"))
			return
		}

		ctx := r.Context()
		accessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			respondError(w, r, fmt.Errorf("visitor_handler.create(): error getting access payload: %w", err))
			return
		}

		if desiredVisitor.ResidentID != "" && desiredVisitor.ResidentID != accessPayload.ID {
			respondError(w, r, BadRequest("Residents cannot create a visitor for another resident"))
			return
		}
		if desiredVisitor.ResidentID == "" {
			desiredVisitor.ResidentID = accessPayload.ID
		}

		visitor, err := h.visitorService.Create(desiredVisitor)
		if err != nil {
			respondError(w, r, err)
			return
		}

		respondJSON(w, r, http.StatusOK, visitor)
	}
}

func (h visitorHandler) deleteOne() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !util.IsUUIDV4(id) {
			respondError(w, r, errs.IDNotUUID)
			return
		}

		ctx := r.Context()
		accessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			respondError(w, r, fmt.Errorf("visitor_handler.deleteOne: error getting access payload: %w", err))
			return
		}

		visitorToDelete, err := h.visitorService.GetOne(id)
		if err != nil {
			respondError(w, r, err)
			return
		}

		if visitorToDelete.ResidentID != accessPayload.ID {
			respondError(w, r, errs.Unauthorized)
			return
		}

		if err := h.visitorService.Delete(id); err != nil {
			respondError(w, r, err)
			return
		}

		respondJSON(w, r, http.StatusOK, message{"Successfully deleted visitor"})
	}
}
