package api

import (
	"errors"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
)

func getAllResidents(residentRepo storage.ResidentRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := toPosInt(r.URL.Query().Get("limit"))
		page := toPosInt(r.URL.Query().Get("page"))
		boundedLimit, offset := getBoundedLimitAndOffset(limit, page)

		allResidents, err := residentRepo.GetAll(boundedLimit, offset)
		if err != nil {
			log.Error().Msgf("resident_router.getAll: Error querying residentRepo: %v", err)
			respondInternalError(w)
			return
		}

		totalAmount, err := residentRepo.GetAllTotalAmount()
		if err != nil {
			log.Error().Msgf("resident_router.getAll: Error getting total amount: %v", err)
			respondInternalError(w)
			return
		}

		residentsWithMetadata := newListWithMetadata(allResidents, totalAmount)

		respondJSON(w, http.StatusOK, residentsWithMetadata)
	}
}

func getOneResident(residentRepo storage.ResidentRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			respondError(w, newErrBadRequest("id parameter cannot be empty"))
			return
		}

		resident, err := residentRepo.GetOne(id)
		if err != nil && !errors.Is(err, storage.ErrNoRows) {
			log.Error().Msgf("resident_router.getOne: Error getting resident: %v", err)
			respondInternalError(w)
			return
		} else if errors.Is(err, storage.ErrNoRows) {
			respondError(w, newErrNotFound("resident"))
			return
		}

		respondJSON(w, http.StatusOK, resident)
	}
}
