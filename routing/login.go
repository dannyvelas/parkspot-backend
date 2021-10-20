package routing

import (
	"encoding/json"
	"fmt"
	"github.com/dannyvelas/parkspot-api/storage"
	"github.com/dannyvelas/parkspot-api/utils"
	"net/http"
)

type credentials struct {
	id       string
	password string
}

func HandleLogin(adminRepo storage.AdminRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds credentials
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		admin, err := adminRepo.GetOne(r.URL.Opaque)
		if err != nil && err == storage.ResourceNotFound {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		} else if err != nil {
			fmt.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}

		utils.RespondJson(w, http.StatusOK, admin)
	}
}
