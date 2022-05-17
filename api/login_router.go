package api

import (
	"encoding/json"
	"errors"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type credentials struct {
	Id       string
	Password string
}

func Login(jwtMiddleware JWTMiddleware, adminRepo storage.AdminRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds credentials
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			respondError(w, newErrMalformed("Credentials"))
			return
		}

		admin, err := adminRepo.GetOne(creds.Id)
		if errors.Is(err, storage.ErrNoRows) {
			respondError(w, errUnauthorized)
			return
		} else if err != nil {
			log.Error().Msgf("login_router: Error querying adminRepo: %v", err)
			respondInternalError(w)
			return
		}

		if err = bcrypt.CompareHashAndPassword(
			[]byte(admin.Password),
			[]byte(creds.Password),
		); err != nil {
			respondError(w, errUnauthorized)
			return
		}

		token, err := jwtMiddleware.newJWT(admin.Id, AdminRole)
		if err != nil {
			log.Error().Msgf("login_router: Error generating JWT: %v", err)
			respondInternalError(w)
			return
		}

		cookie := http.Cookie{Name: "jwt", Value: token, HttpOnly: true, Path: "/"}
		http.SetCookie(w, &cookie)

		respondJSON(w, http.StatusOK, admin)
	}
}
