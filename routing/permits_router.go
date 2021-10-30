package routing

import (
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
		page := ToUint(r.URL.Query().Get("page"))
		size := ToUint(r.URL.Query().Get("size"))
		limit, offset := PagingToLimitOffset(page, size)

		activePermits, err := permitRepo.GetActive(limit, offset)
		if err != nil {
			HandleInternalError(w, "Error querying permitRepo: "+err.Error())
			return
		}

		RespondJson(w, http.StatusOK, activePermits)
	}
}
