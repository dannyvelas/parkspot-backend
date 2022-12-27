package api

import (
	"encoding/json"
	"github.com/dannyvelas/lasvistas_api/app"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
	"time"
)

type authHandler struct {
	jwtService  app.JWTService
	authService app.AuthService
}

func newAuthHandler(jwtService app.JWTService, authService app.AuthService) authHandler {
	return authHandler{
		jwtService:  jwtService,
		authService: authService,
	}
}

func (h authHandler) login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var credentials struct {
			ID       string
			Password string
		}
		err := json.NewDecoder(r.Body).Decode(&credentials)
		if err != nil {
			respondError(w, *errs.Malformed("Credentials"))
			return
		}

		session, refreshToken, apiErr := h.authService.Login(credentials.ID, credentials.Password)
		if apiErr != nil {
			respondError(w, *apiErr)
			return
		}

		h.sendRefreshToken(w, refreshToken)

		respondJSON(w, http.StatusOK, session)
	}
}

func (h authHandler) logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie := http.Cookie{Name: config.RefreshCookieKey, Value: "deleted", HttpOnly: true, Path: "/", Expires: time.Unix(0, 0)}
		http.SetCookie(w, &cookie)

		respondJSON(w, http.StatusOK, message{"Successfully logged-out user"})
	}
}

func (h authHandler) refreshTokens() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(config.RefreshCookieKey)
		if err != nil {
			respondError(w, *errs.Unauthorized)
			return
		}

		refreshPayload, err := h.jwtService.ParseRefresh(cookie.Value)
		if err != nil {
			respondError(w, *errs.Unauthorized)
			return
		}

		session, refreshToken, apiErr := h.authService.RefreshTokens(refreshPayload)
		if apiErr != nil {
			respondError(w, *apiErr)
			return
		}

		h.sendRefreshToken(w, refreshToken)

		respondJSON(w, http.StatusOK, session)
	}
}

func (h authHandler) sendResetPasswordEmail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		emailSentResponse := message{"If this account is in our database, instructions to" +
			" reset a password have been sent to the email associated with this account."}

		var payload struct{ ID string }
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			respondError(w, *errs.Malformed("id object"))
			return
		} else if payload.ID == "" {
			respondError(w, *errs.EmptyFields("id"))
			return
		}

		if apiErr := h.authService.SendResetPasswordEmail(r.Context(), payload.ID); apiErr != nil {
			respondError(w, *apiErr)
			return
		}

		respondJSON(w, http.StatusOK, emailSentResponse)
	}
}

func (h authHandler) resetPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct{ Password string }
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			respondError(w, *errs.Malformed("password object"))
			return
		} else if payload.Password == "" {
			respondError(w, *errs.EmptyFields("password"))
			return
		}

		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Debug().Msg("No 'Authorization' header was present with 'Bearer ' prefix.")
			respondError(w, *errs.Unauthorized)
			return
		}

		accessToken := strings.TrimPrefix(authHeader, "Bearer ")
		user, err := h.jwtService.ParseAccess(accessToken)
		if err != nil {
			respondError(w, *errs.Unauthorized)
			return
		}

		if apiErr := h.authService.ResetPassword(user.ID, payload.Password); apiErr != nil {
			respondError(w, *apiErr)
			return
		}

		respondJSON(w, http.StatusOK, message{"Password has been successfully reset."})
	}
}

func (h authHandler) sendRefreshToken(w http.ResponseWriter, refreshToken string) {
	cookie := http.Cookie{
		Name:     config.RefreshCookieKey,
		Value:    refreshToken,
		HttpOnly: true,
		Path:     "/"}
	http.SetCookie(w, &cookie)
}
