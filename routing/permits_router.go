package routing

import (
	"github.com/dannyvelas/lasvistas_api/routing/internal"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
)

func PermitsRouter(permitRepo storage.PermitRepo) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/active", GetActive(permitRepo))
		r.Get("/all", GetAll(permitRepo))
	}
}

func GetActive(permitRepo storage.PermitRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Msg("Get Active Endpoint")

		size := internal.ToUint(r.URL.Query().Get("size"))
		page := internal.ToUint(r.URL.Query().Get("page"))
		boundedSize, offset := internal.GetBoundedSizeAndOffset(size, page)

		activePermits, err := permitRepo.GetActive(boundedSize, offset)
		if err != nil {
			internal.HandleInternalError(w, "Error querying permitRepo: "+err.Error())
			return
		}

		internal.RespondJson(w, http.StatusOK, activePermits)
	}
}

func GetAll(permitRepo storage.PermitRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Msg("Get All Endpoint")

		size := internal.ToUint(r.URL.Query().Get("size"))
		page := internal.ToUint(r.URL.Query().Get("page"))
		boundedSize, offset := internal.GetBoundedSizeAndOffset(size, page)

		allPermits, err := permitRepo.GetAll(boundedSize, offset)
		if err != nil {
			internal.HandleInternalError(w, "Error querying permitRepo: "+err.Error())
			return
		}

		internal.RespondJson(w, http.StatusOK, allPermits)
	}
}
