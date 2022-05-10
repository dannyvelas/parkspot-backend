package api

import (
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/rs/zerolog/log"
	"net/http"
)

func getAllResidents(residentRepo storage.ResidentRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		size := toUint(r.URL.Query().Get("size"))
		page := toUint(r.URL.Query().Get("page"))
		boundedSize, offset := getBoundedSizeAndOffset(size, page)

		allResidents, err := residentRepo.GetAll(boundedSize, offset)
		if err != nil {
			log.Error().Msgf("resident_router.getAll: Error querying residentRepo: %v", err)
			respondError(w, errInternalServerError)
			return
		}

		respondJSON(w, http.StatusOK, allResidents)
	}
}
