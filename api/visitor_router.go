package api

import (
	"encoding/json"
	"errors"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
)

func getAllVisitors(visitorRepo storage.VisitorRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := toPosInt(r.URL.Query().Get("limit"))
		page := toPosInt(r.URL.Query().Get("page"))
		boundedLimit, offset := getBoundedLimitAndOffset(limit, page)

		allVisitors, err := visitorRepo.GetAll(boundedLimit, offset)
		if err != nil {
			log.Error().Msgf("visitor_router.getAll: Error querying visitorRepo: %v", err)
			respondInternalError(w)
			return
		}

		totalAmount, err := visitorRepo.GetAllTotalAmount()
		if err != nil {
			log.Error().Msgf("visitor_router.getAll: Error getting total amount: %v", err)
			respondInternalError(w)
			return
		}

		visitorsWithMetadata := newListWithMetadata(allVisitors, totalAmount)

		respondJSON(w, http.StatusOK, visitorsWithMetadata)
	}
}

func searchVisitors(visitorRepo storage.VisitorRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		searchStr := r.URL.Query().Get("search")
		if searchStr == "" {
			respondJSON(w, http.StatusOK, []models.Visitor{})
			return
		}

		visitors, err := visitorRepo.Search(searchStr)
		if err != nil {
			log.Error().Msgf("visitor_router.searchVisitors: Error getting visitors: %v", err)
			respondInternalError(w)
			return
		}

		visitorsWithMetadata := newListWithMetadata(visitors, len(visitors))

		respondJSON(w, http.StatusOK, visitorsWithMetadata)
	}
}

func getVisitorsOfResident(visitorRepo storage.VisitorRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, err := ctxGetUser(ctx)
		if err != nil {
			log.Error().Msgf("visitor_router.getVisitorsOfResident: %v", err)
			respondInternalError(w)
			return
		}

		visitors, err := visitorRepo.GetOfResident(user.Id)
		if err != nil {
			log.Error().Msgf("visitor_router.getVisitorsOfResident: Error querying visitorRepo: %v", err)
			respondInternalError(w)
			return
		}

		visitorsWithMetadata := newListWithMetadata(visitors, len(visitors))

		respondJSON(w, http.StatusOK, visitorsWithMetadata)
	}
}

func createVisitor(visitorRepo storage.VisitorRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, err := ctxGetUser(ctx)
		if err != nil {
			log.Error().Msgf("visitor_router.getVisitorsOfResident: %v", err)
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

		visitorId, err := visitorRepo.Create(user.Id,
			payload.FirstName,
			payload.LastName,
			payload.Relationship,
			payload.AccessStart,
			payload.AccessEnd)
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
