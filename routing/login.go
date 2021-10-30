package routing

import (
	"encoding/json"
	"errors"
	"github.com/dannyvelas/parkspot-api/auth"
	"github.com/dannyvelas/parkspot-api/storage"
	"github.com/dannyvelas/parkspot-api/utils"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type credentials struct {
	Id       string
	Password string
}

func Login(authenticator auth.Authenticator, adminRepo storage.AdminRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds credentials
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			utils.HandleError(w, utils.BadRequest)
			return
		}

		admin, err := adminRepo.GetOne(creds.Id)
		if errors.As(err, &storage.NotFound{}) {
			utils.HandleError(w, utils.Unauthorized)
			return
		} else if err != nil {
			utils.HandleInternalError(w, "Error querying adminRepo: "+err.Error())
			return
		}

		err = bcrypt.CompareHashAndPassword(
			[]byte(admin.Password),
			[]byte(creds.Password),
		)
		if err != nil {
			utils.HandleError(w, utils.Unauthorized)
			return
		}

		token, err := authenticator.NewJWT(admin.Id)
		if err != nil {
			utils.HandleInternalError(w, "Error generating JWT: "+err.Error())
			return
		}

		cookie := http.Cookie{Name: "jwt", Value: token, HttpOnly: true, Path: "/"}
		http.SetCookie(w, &cookie)
	}
}
