package routing

import (
	"encoding/json"
	"errors"
	"github.com/dannyvelas/parkspot-api/auth"
	"github.com/dannyvelas/parkspot-api/routing/internal"
	"github.com/dannyvelas/parkspot-api/storage"
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
			internal.HandleError(w, internal.BadRequest)
			return
		}

		admin, err := adminRepo.GetOne(creds.Id)
		if errors.As(err, &storage.NotFound{}) {
			internal.HandleError(w, internal.Unauthorized)
			return
		} else if err != nil {
			internal.HandleInternalError(w, "Error querying adminRepo: "+err.Error())
			return
		}

		err = bcrypt.CompareHashAndPassword(
			[]byte(admin.Password),
			[]byte(creds.Password),
		)
		if err != nil {
			internal.HandleError(w, internal.Unauthorized)
			return
		}

		token, err := authenticator.NewJWT(admin.Id)
		if err != nil {
			internal.HandleInternalError(w, "Error generating JWT: "+err.Error())
			return
		}

		cookie := http.Cookie{Name: "jwt", Value: token, HttpOnly: true, Path: "/"}
		http.SetCookie(w, &cookie)
	}
}
