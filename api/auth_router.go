package api

import (
	"encoding/json"
	"errors"
	"github.com/dannyvelas/lasvistas_api/app"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
	"time"
)

type AuthHandler struct {
	jwtService  app.JWTService
	authService app.AuthService
}

func NewAuthHandler(jwtService app.JWTService, authService app.AuthService) AuthHandler {
	return AuthHandler{
		jwtService:  jwtService,
		authService: authService,
	}
}

func (h AuthHandler) login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var credentials struct {
			ID       string
			Password string
		}
		err := json.NewDecoder(r.Body).Decode(&credentials)
		if err != nil {
			respondError(w, newErrMalformed("Credentials"))
			return
		}

		session, refreshToken, err := h.authService.Login(credentials.ID, credentials.Password)
		if errors.Is(err, app.ErrUnauthorized) {
			respondError(w, errUnauthorized)
			return
		} else if err != nil {
			log.Error().Msgf("auth_router.login: %v", err)
			respondInternalError(w)
			return
		}

		h.sendRefreshToken(w, refreshToken)

		respondJSON(w, http.StatusOK, session)
	}
}

func (h AuthHandler) logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie := http.Cookie{Name: config.RefreshCookieKey, Value: "deleted", HttpOnly: true, Path: "/", Expires: time.Unix(0, 0)}
		http.SetCookie(w, &cookie)

		respondJSON(w, http.StatusOK, message{"Successfully logged-out user"})
	}
}

func (h AuthHandler) refreshTokens() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(config.RefreshCookieKey)
		if err != nil {
			respondError(w, errUnauthorized)
			return
		}

		refreshPayload, err := h.jwtService.ParseRefresh(cookie.Value)
		if err != nil {
			respondError(w, errUnauthorized)
			return
		}

		session, refreshToken, err := h.authService.RefreshTokens(refreshPayload)
		if errors.Is(err, app.ErrUnauthorized) {
			respondError(w, errUnauthorized)
			return
		} else if err != nil {
			log.Error().Msg("auth_router.refreshTokens: " + err.Error())
			respondInternalError(w)
			return
		}

		h.sendRefreshToken(w, refreshToken)

		respondJSON(w, http.StatusOK, session)
	}
}

func (h AuthHandler) sendResetPasswordEmail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		emailSentResponse := message{"If this account is in our database, instructions to" +
			" reset a password have been sent to the email associated with this account."}

		var payload struct{ ID string }
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			respondError(w, newErrMalformed("id object"))
			return
		} else if payload.ID == "" {
			respondError(w, errEmptyFields)
			return
		}

		err := h.authService.SendResetPasswordEmail(r.Context(), payload.ID)
		if errors.Is(err, app.ErrUnauthorized) {
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

func (h AuthHandler) resetPassword() http.HandlerFunc {
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
		user, err := h.jwtService.ParseAccess(accessToken)
		if err != nil {
			log.Debug().Msgf("Error parsing: %v", err)
			respondError(w, errUnauthorized)
			return
		}

		err = h.authService.ResetPassword(user.ID, payload.Password)
		if err != nil {
			log.Error().Msgf("auth_router.resetPassword: error calling service: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, message{"Password has been successfully reset."})
	}
}

func (h AuthHandler) sendRefreshToken(w http.ResponseWriter, refreshToken string) {
	cookie := http.Cookie{
		Name:     config.RefreshCookieKey,
		Value:    refreshToken,
		HttpOnly: true,
		Path:     "/"}
	http.SetCookie(w, &cookie)
}
