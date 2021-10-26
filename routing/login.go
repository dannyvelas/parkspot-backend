package routing

import (
	"encoding/json"
	"errors"
	"github.com/dannyvelas/parkspot-api/storage"
	"github.com/dannyvelas/parkspot-api/utils"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type credentials struct {
	Id       string
	Password string
}

func Login(adminRepo storage.AdminRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds credentials
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		admin, err := adminRepo.GetOne(creds.Id)
		if errors.As(err, &storage.NotFound{}) {
			http.Error(w, "Wrong Credentials", http.StatusUnauthorized)
			return
		} else if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}

		err = bcrypt.CompareHashAndPassword(
			[]byte(admin.Password),
			[]byte(creds.Password),
		)
		if err != nil {
			http.Error(w, "Wrong Credentials", http.StatusUnauthorized)
			return
		}

		utils.RespondJson(w, http.StatusOK, admin)
	}
}
