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
	"time"
)

type VisitorHandler struct {
	visitorService app.VisitorService
}

func NewVisitorHandler(visitorService app.VisitorService) VisitorHandler {
	return VisitorHandler{
		visitorService: visitorService,
	}
}

func (h VisitorHandler) GetActive() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := util.ToPosInt(r.URL.Query().Get("limit"))
		page := util.ToPosInt(r.URL.Query().Get("page"))
		search := r.URL.Query().Get("search")

		ctx := r.Context()
		AccessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			log.Error().Msgf("visitor_router.getVisitorsOfResident: %v", err)
			respondInternalError(w)
			return
		}

		residentID := ""
		if AccessPayload.Role == models.ResidentRole {
			residentID = AccessPayload.ID
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

func (h VisitorHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		AccessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			log.Error().Msgf("error getting access payload in visitor handler: %v", err)
			respondInternalError(w)
			return
		}

		var payload newVisitorReq
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			respondError(w, newErrMalformed("NewVisitorReq"))
			return
		}

		if err := payload.validate(); err != nil {
			respondError(w, newErrBadRequest(err.Error()))
			return
		}

		startTS, endTS := payload.AccessStart.Unix(), payload.AccessEnd.Unix()
		if payload.IsForever {
			startTS = time.Now().Unix()
			endTS = models.EndOfTime.Unix()
		}

		desiredVisitor := models.CreateVisitor{
			ResidentID:   AccessPayload.ID,
			FirstName:    payload.FirstName,
			LastName:     payload.LastName,
			Relationship: payload.Relationship,
			StartTS:      startTS,
			EndTS:        endTS,
		}

		visitor, err := h.visitorService.Create(desiredVisitor)
		if err != nil {
			log.Error().Msgf("error creating visitor in visitor service: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, visitor)
	}
}

func (h VisitorHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !util.IsUUIDV4(id) {
			respondError(w, newErrBadRequest("id parameter is not a UUID"))
			return
		}

		err := h.visitorService.Delete(id)
		if errors.Is(err, app.ErrNotFound) {
			respondError(w, newErrNotFound("visitor"))
			return
		} else if err != nil {
			log.Error().Msgf("error deleting visitor in visitor service: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, message{"Successfully deleted visitor"})
	}
}
