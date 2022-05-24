package api

import (
	"github.com/dannyvelas/lasvistas_api/storage"
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
