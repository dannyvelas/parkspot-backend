package api

import (
	"encoding/json"
	"errors"
	"github.com/dannyvelas/lasvistas_api/services"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
	"time"
)

func login(authService services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var credentials struct {
			Id       string
			Password string
		}
		err := json.NewDecoder(r.Body).Decode(&credentials)
		if err != nil {
			respondError(w, newErrMalformed("Credentials"))
			return
		}

		session, refreshToken, err := authService.Login(credentials.Id, credentials.Password)
		if errors.Is(err, services.ErrUnauthorized) {
			respondError(w, errUnauthorized)
			return
		} else if err != nil {
			log.Error().Msgf("auth_router.login: %v", err)
			respondInternalError(w)
			return
		}

		sendRefreshToken(w, refreshToken)

		respondJSON(w, http.StatusOK, session)
	}
}

func logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie := http.Cookie{Name: refreshCookieKey, Value: "deleted", HttpOnly: true, Path: "/", Expires: time.Unix(0, 0)}
		http.SetCookie(w, &cookie)

		respondJSON(w, http.StatusOK, message{"Successfully logged-out user"})
	}
}

func refreshTokens(jwtService services.JWTService, authService services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(refreshCookieKey)
		if err != nil {
			respondError(w, errUnauthorized)
			return
		}

		refreshPayload, err := jwtService.ParseRefresh(cookie.Value)
		if err != nil {
			respondError(w, errUnauthorized)
			return
		}

		session, refreshToken, err := authService.RefreshTokens(refreshPayload)
		if errors.Is(err, services.ErrUnauthorized) {
			respondError(w, errUnauthorized)
			return
		} else if err != nil {
			log.Error().Msg("auth_router.refreshTokens: " + err.Error())
			respondInternalError(w)
			return
		}

		sendRefreshToken(w, refreshToken)

		respondJSON(w, http.StatusOK, session)
	}
}

func sendResetPasswordEmail(jwtService services.JWTService, authService services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		emailSentResponse := message{"If this account is in our database, instructions to" +
			" reset a password have been sent to the email associated with this account."}

		var payload struct{ Id string }
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			respondError(w, newErrMalformed("id object"))
			return
		} else if payload.Id == "" {
			respondError(w, errEmptyFields)
			return
		}

		err := authService.SendResetPasswordEmail(r.Context(), payload.Id)
		if errors.Is(err, services.ErrUnauthorized) {
			respondError(w, errUnauthorized)
			return
		} else if err != nil {
			log.Error().Msg("auth_router.sendResetPasswordEmail: " + err.Error())
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, emailSentResponse)
	}
}

func resetPassword(jwtService services.JWTService, authService services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct{ Password string }
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			respondError(w, newErrMalformed("password object"))
			return
		} else if payload.Password == "" {
			respondError(w, errEmptyFields)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Debug().Msg("No 'Authorization' header was present with 'Bearer ' prefix.")
			respondError(w, errUnauthorized)
			return
		}

		accessToken := strings.TrimPrefix(authHeader, "Bearer ")
		user, err := jwtService.ParseAccess(accessToken)
		if err != nil {
			log.Debug().Msgf("Error parsing: %v", err)
			respondError(w, errUnauthorized)
			return
		}

		err = authService.ResetPassword(user.Id, payload.Password)
		if err != nil {
			log.Error().Msgf("auth_router.resetPassword: error calling service: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, message{"Password has been successfully reset."})
	}
}

func sendRefreshToken(w http.ResponseWriter, refreshToken string) {
	cookie := http.Cookie{
		Name:     refreshCookieKey,
		Value:    refreshToken,
		HttpOnly: true,
		Path:     "/"}
	http.SetCookie(w, &cookie)
}
