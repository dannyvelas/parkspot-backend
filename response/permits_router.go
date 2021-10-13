package response

import (
	"github.com/dannyvelas/parkspot-api/storage"
	"net/http"
)

func PermitsRouter(permitRepo storage.PermitRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.GET("/active", GetActivePermits(permitRepo))
	}
}

func GetActivePermits(permitRepo storage.PermitRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		activePermits, err := permitRepo.GetActive()

		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}

		respondJson(activePermits)
	}
}
