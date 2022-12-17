package api

import (
	"encoding/json"
	"errors"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

func getActiveVisitors(visitorRepo storage.VisitorRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := toPosInt(r.URL.Query().Get("limit"))
		page := toPosInt(r.URL.Query().Get("page"))
		search := r.URL.Query().Get("search")
		boundedLimit, offset := getBoundedLimitAndOffset(limit, page)

		ctx := r.Context()
		AccessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			log.Error().Msgf("visitor_router.getVisitorsOfResident: %v", err)
			respondInternalError(w)
			return
		}

		residentID := ""
		if AccessPayload.Role == models.ResidentRole {
			residentID = AccessPayload.Id
		}

		allVisitors, err := visitorRepo.Get(true, residentID, search, boundedLimit, offset)
		if err != nil {
			log.Error().Msgf("visitor_router.getActiveVisitors: Error querying visitorRepo: %v", err)
			respondInternalError(w)
			return
		}

		totalAmount, err := visitorRepo.GetCount(true, residentID)
		if err != nil {
			log.Error().Msgf("visitor_router.getActiveVisitors: Error getting total amount: %v", err)
			respondInternalError(w)
			return
		}

		visitorsWithMetadata := newListWithMetadata(allVisitors, totalAmount)

		respondJSON(w, http.StatusOK, visitorsWithMetadata)
	}
}

func createVisitor(visitorRepo storage.VisitorRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		AccessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			log.Error().Msgf("visitor_router.createVisitor: error getting access payload: %v", err)
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

		visitorId, err := visitorRepo.Create(
			AccessPayload.Id,
			payload.FirstName,
			payload.LastName,
			payload.Relationship,
			startTS,
			endTS)
		if err != nil {
			log.Error().Msgf("visitor_router.createVisitor: Error creating visitor: %v", err)
			respondInternalError(w)
			return
		}

		visitor, err := visitorRepo.GetOne(visitorId)
		if err != nil {
			log.Error().Msgf("visitor_router.createVisitor: Error getting visitor: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, visitor)
	}
}

func deleteVisitor(visitorRepo storage.VisitorRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !isUUIDV4(id) {
			respondError(w, newErrBadRequest("id parameter is not a UUID"))
			return
		}

		err := visitorRepo.Delete(id)
		if errors.Is(err, storage.ErrNoRows) {
			respondError(w, newErrNotFound("visitor"))
			return
		} else if err != nil {
			log.Error().Msgf("visitor_router.deleteVisitor: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, message{"Successfully deleted visitor"})
	}
}
