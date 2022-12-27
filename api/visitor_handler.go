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
			log.Error().Msgf("visitor_router.getVisitorsOfResident: %v", err)
			respondInternalError(w)
			return
		}

		residentID := ""
		if accessPayload.Role == models.ResidentRole {
			residentID = accessPayload.ID
		}

		visitorsWithMetadata, err := h.visitorService.GetActive(limit, page, search, residentID)
		if err != nil {
			log.Error().Msgf("visitor_router.getActiveVisitors: Error querying visitorRepo: %v", err)
			respondInternalError(w)
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
			log.Error().Msgf("error getting access payload in visitor handler: %v", err)
			respondInternalError(w)
			return
		}

		var payload models.Visitor
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			respondError(w, newErrMalformed("NewVisitorReq"))
			return
		}

		if err := payload.ValidateCreation(); err != nil {
			respondError(w, newErrBadRequest(err.Error()))
			return
		}

		visitor, err := h.visitorService.Create(accessPayload.ID, payload)
		if err != nil {
			log.Error().Msgf("error creating visitor in visitor service: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, visitor)
	}
}

func (h visitorHandler) deleteOne() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !util.IsUUIDV4(id) {
			respondError(w, newErrBadRequest("id parameter is not a UUID"))
			return
		}

		err := h.visitorService.Delete(id)
		var apiErr errs.ApiErr
		if errors.Is(err, errs.NotFound) {
			respondError(w, newErrNotFound("visitor"))
			return
		} else if errors.As(err, &apiErr) {
			respondError(w, newErrBadRequest(err.Error()))
			return
		} else if err != nil {
			log.Error().Msgf("error deleting visitor in visitor service: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, message{"Successfully deleted visitor"})
	}
}
