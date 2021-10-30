package routing

import (
	"github.com/dannyvelas/parkspot-api/routing/internal"
	"github.com/dannyvelas/parkspot-api/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func PermitsRouter(permitRepo storage.PermitRepo) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/active", GetActive(permitRepo))
	}
}

func GetActive(permitRepo storage.PermitRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := internal.ToUint(r.URL.Query().Get("page"))
		size := internal.ToUint(r.URL.Query().Get("size"))
		limit, offset := internal.PagingToLimitOffset(page, size)

		activePermits, err := permitRepo.GetActive(limit, offset)
		if err != nil {
			internal.HandleInternalError(w, "Error querying permitRepo: "+err.Error())
			return
		}

		internal.RespondJson(w, http.StatusOK, activePermits)
	}
}
