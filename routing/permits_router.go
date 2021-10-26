package routing

import (
	"github.com/dannyvelas/parkspot-api/storage"
	"github.com/dannyvelas/parkspot-api/utils"
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
		page := utils.ToUint(r.URL.Query().Get("page"))
		size := utils.ToUint(r.URL.Query().Get("size"))

		limit, offset := utils.PagingToLimitOffset(page, size)
		activePermits, err := permitRepo.GetActive(limit, offset)

		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}

		utils.RespondJson(w, http.StatusOK, activePermits)
	}
}
