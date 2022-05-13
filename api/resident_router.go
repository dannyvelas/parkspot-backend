package api

import (
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/rs/zerolog/log"
	"net/http"
)

func getAllResidents(residentRepo storage.ResidentRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := toUint(r.URL.Query().Get("limit"))
		page := toUint(r.URL.Query().Get("page"))
		boundedLimit, offset := getBoundedLimitAndOffset(limit, page)

		allResidents, err := residentRepo.GetAll(boundedLimit, offset)
		if err != nil {
			log.Error().Msgf("resident_router.getAll: Error querying residentRepo: %v", err)
			respondError(w, errInternalServerError)
			return
		}

		respondJSON(w, http.StatusOK, allResidents)
	}
}
