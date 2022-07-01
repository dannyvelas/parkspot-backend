package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"regexp"
	"time"
)

type credentials struct {
	Id       string
	Password string
}

func login(jwtMiddleware jwtMiddleware, adminRepo storage.AdminRepo, residentRepo storage.ResidentRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds credentials
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			respondError(w, newErrMalformed("Credentials"))
			return
		}

		user, err := getUser(creds.Id, creds.Password, adminRepo, residentRepo)
		if err != nil && errors.Is(err, errUnauthorized) {
			respondError(w, errUnauthorized)
			return
		} else if err != nil {
			log.Error().Msgf("auth_router: Error getting: %v", err)
			respondInternalError(w)
			return
		}

		token, err := jwtMiddleware.newJWT(user.Id, user.FirstName, user.LastName, user.Email, user.Role)
		if err != nil {
			log.Error().Msgf("auth_router: Error generating JWT: %v", err)
			respondInternalError(w)
			return
		}

		cookie := http.Cookie{Name: "jwt", Value: token, HttpOnly: true, Path: "/"}
		http.SetCookie(w, &cookie)

		respondJSON(w, http.StatusOK, user)
	}
}

func logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie := http.Cookie{Name: "jwt", Value: "deleted", HttpOnly: true, Path: "/", Expires: time.Unix(0, 0)}
		http.SetCookie(w, &cookie)

		respondJSON(w, http.StatusOK, message{"Successfully logged-out user"})
	}
}

func getMe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, err := ctxGetUser(ctx)
		if err != nil {
			log.Error().Msgf("auth_router.getMe: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, user)
	}
}

// helpers
func getUser(username, password string, adminRepo storage.AdminRepo, residentRepo storage.ResidentRepo) (user, error) {
	var userFound user
	var hash string

	if !regexp.MustCompile("^(B|T)\\d{7}$").MatchString(username) {
		admin, err := adminRepo.GetOne(username)
		if errors.Is(err, storage.ErrNoRows) {
			return user{}, errUnauthorized
		} else if err != nil {
			return user{}, fmt.Errorf("Error querying adminRepo: %v", err)
		}

		userFound = newUser(admin.Id, admin.FirstName, admin.LastName, admin.Email, AdminRole)
		hash = admin.Password
	} else {
		resident, err := residentRepo.GetOne(username)
		if errors.Is(err, storage.ErrNoRows) {
			return user{}, errUnauthorized
		} else if err != nil {
			return user{}, fmt.Errorf("Error querying residentRepo: %v", err)
		}

		userFound = newUser(resident.Id, resident.FirstName, resident.LastName, resident.Email, ResidentRole)
		hash = resident.Password
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	); err != nil {
		return user{}, errUnauthorized
	}

	return userFound, nil
}
