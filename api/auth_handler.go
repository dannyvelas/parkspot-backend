package api

import (
	"encoding/json"
	"github.com/dannyvelas/parkspot-backend/app"
	"github.com/dannyvelas/parkspot-backend/config"
	"github.com/dannyvelas/parkspot-backend/errs"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
	"time"
)

type authHandler struct {
	httpConfig  config.HttpConfig
	jwtService  app.JWTService
	authService app.AuthService
}

func newAuthHandler(httpConfig config.HttpConfig, jwtService app.JWTService, authService app.AuthService) authHandler {
	return authHandler{
		httpConfig:  httpConfig,
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
			respondError(w, errs.Malformed("Credentials"))
			return
		}

		session, refreshToken, err := h.authService.Login(credentials.ID, credentials.Password)
		if err != nil {
			respondError(w, err)
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
			log.Debug().Msgf("could not find cookie")
			respondError(w, errs.Unauthorized)
			return
		}

		refreshPayload, err := h.jwtService.ParseRefresh(cookie.Value)
		if err != nil {
			log.Debug().Msgf("could not parse cookie: %v", err)
			respondError(w, errs.Unauthorized)
			return
		}

		session, refreshToken, err := h.authService.RefreshTokens(refreshPayload)
		if err != nil {
			log.Debug().Msgf("could not have auth service refresh tokens: %v", err)
			respondError(w, err)
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
			respondError(w, errs.Malformed("id object"))
			return
		} else if payload.ID == "" {
			respondError(w, errs.EmptyFields("id"))
			return
		}

		if err := h.authService.SendResetPasswordEmail(r.Context(), payload.ID); err != nil {
			respondError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, emailSentResponse)
	}
}

func (h authHandler) resetPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct{ Password string }
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			respondError(w, errs.Malformed("password object"))
			return
		}

		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Debug().Msg("No 'Authorization' header was present with 'Bearer ' prefix.")
			respondError(w, errs.Unauthorized)
			return
		}

		accessToken := strings.TrimPrefix(authHeader, "Bearer ")
		user, err := h.jwtService.ParseAccess(accessToken)
		if err != nil {
			respondError(w, errs.Unauthorized)
			return
		}

		if err := h.authService.ResetPassword(user.ID, payload.Password); err != nil {
			respondError(w, err)
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
		Path:     "/",
		Domain:   h.httpConfig.CookieDomain,
	}
	http.SetCookie(w, &cookie)
}
