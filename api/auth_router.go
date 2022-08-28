package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type authResponse struct {
	User        user   `json:"user"`
	AccessToken string `json:"accessToken"`
}

func login(jwtMiddleware jwtMiddleware, adminRepo storage.AdminRepo, residentRepo storage.ResidentRepo) http.HandlerFunc {
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

		userFound, hash, err := getUserAndHashById(credentials.Id, adminRepo, residentRepo)
		if errors.Is(err, errUnauthorized) {
			respondError(w, errUnauthorized)
			return
		} else if err != nil {
			log.Error().Msg("auth_router.login: " + err.Error())
			respondInternalError(w)
			return
		}

		if err := bcrypt.CompareHashAndPassword(
			[]byte(hash),
			[]byte(credentials.Password),
		); err != nil {
			respondError(w, errUnauthorized)
			return
		}

		refreshToken, err := jwtMiddleware.newRefresh(userFound.Id, userFound.TokenVersion)
		if err != nil {
			log.Error().Msgf("auth_router.login: Error generating refresh JWT: %v", err)
			respondInternalError(w)
			return
		}

		sendRefreshToken(w, refreshToken)

		accessToken, err := jwtMiddleware.newAccess(userFound.Id, userFound.Role)
		if err != nil {
			log.Error().Msgf("auth_router.login: Error generating access JWT: %v", err)
			respondInternalError(w)
			return
		}

		response := authResponse{userFound, accessToken}

		respondJSON(w, http.StatusOK, response)
	}
}

func logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie := http.Cookie{Name: refreshCookieKey, Value: "deleted", HttpOnly: true, Path: "/", Expires: time.Unix(0, 0)}
		http.SetCookie(w, &cookie)

		respondJSON(w, http.StatusOK, message{"Successfully logged-out user"})
	}
}

func refreshTokens(jwtMiddleware jwtMiddleware, adminRepo storage.AdminRepo, residentRepo storage.ResidentRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(refreshCookieKey)
		if err != nil {
			respondError(w, errUnauthorized)
			return
		}

		refreshPayload, err := jwtMiddleware.parseRefresh(cookie.Value)
		if err != nil {
			respondError(w, errUnauthorized)
			return
		}

		user, _, err := getUserAndHashById(refreshPayload.Id, adminRepo, residentRepo)
		if errors.Is(err, errUnauthorized) {
			respondError(w, errUnauthorized)
			return
		} else if err != nil {
			log.Error().Msg("auth_router.getNewTokens: " + err.Error())
			respondInternalError(w)
			return
		}

		if user.TokenVersion != refreshPayload.Version {
			respondError(w, errUnauthorized)
			return
		}

		// refresh the refresh token
		refreshToken, err := jwtMiddleware.newRefresh(user.Id, user.TokenVersion)
		if err != nil {
			log.Error().Msgf("auth_router.getNewTokens: Error generating refresh JWT: %v", err)
			respondInternalError(w)
			return
		}

		sendRefreshToken(w, refreshToken)

		// refresh access token
		accessToken, err := jwtMiddleware.newAccess(user.Id, user.Role)
		if err != nil {
			log.Error().Msgf("auth_router.getNewTokens: Error generating access JWT: %v", err)
			respondInternalError(w)
			return
		}

		response := authResponse{user, accessToken}

		respondJSON(w, http.StatusOK, response)
	}
}

func createResident(residentRepo storage.ResidentRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload newResidentReq
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			respondError(w, newErrMalformed("NewResidentReq"))
			return
		}

		if err := payload.validate(); err != nil {
			respondError(w, newErrBadRequest(err.Error()))
			return
		}

		if _, err := residentRepo.GetOne(payload.ResidentId); err == nil {
			respondError(w, newErrBadRequest("Resident with this id already exists. Please delete the old account if necessary."))
			return
		} else if !errors.Is(err, storage.ErrNoRows) {
			log.Error().Msgf("auth_router.createResident: error getting resident by id: %v", err)
			respondInternalError(w)
			return
		}

		if _, err := residentRepo.GetOneByEmail(payload.Email); err == nil {
			respondError(w, newErrBadRequest("Resident with this email already exists. Please delete the old account or use a different email."))
			return
		} else if !errors.Is(err, storage.ErrNoRows) {
			log.Error().Msgf("auth_router.createResident error getting resident by email: %v", err)
			respondInternalError(w)
			return
		}

		hashBytes, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Error().Msg("auth_router.createResident: error generating hash:" + err.Error())
			respondInternalError(w)
			return
		}
		hashString := string(hashBytes)

		err = residentRepo.Create(payload.ResidentId,
			payload.FirstName,
			payload.LastName,
			payload.Phone,
			payload.Email,
			hashString,
			payload.UnlimDays)
		if err != nil {
			log.Error().Msgf("auth_router.createResident: Error querying residentRepo: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, message{"Resident successfully created."})
	}
}

func sendResetPasswordEmail(jwtMiddleware jwtMiddleware, oauthConfig config.OAuthConfig, adminRepo storage.AdminRepo, residentRepo storage.ResidentRepo) http.HandlerFunc {
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

		userFound, _, err := getUserAndHashById(payload.Id, adminRepo, residentRepo)
		if errors.Is(err, errUnauthorized) {
			respondJSON(w, http.StatusOK, emailSentResponse)
			return
		} else if err != nil {
			log.Error().Msg("auth_router.sendResetPasswordEmail: " + err.Error())
			respondInternalError(w)
			return
		}

		service, err := getGmailService(r.Context(), oauthConfig)
		if err != nil {
			log.Error().Msgf("auth_router.sendResetPasswordEmail: " + err.Error())
			respondInternalError(w)
			return
		}

		gmailMessage, err := createGmailMessage(jwtMiddleware, userFound)
		if err != nil {
			log.Error().Msgf("auth_router.sendResetPasswordEmail: " + err.Error())
			respondInternalError(w)
			return
		}

		_, err = service.Users.Messages.Send("me", gmailMessage).Do()
		if err != nil {
			log.Error().Msg("auth_router.sendResetPasswordEmail: error sending mail:" + err.Error())
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, emailSentResponse)
	}
}

func resetPassword(jwtMiddleware jwtMiddleware, adminRepo storage.AdminRepo, residentRepo storage.ResidentRepo) http.HandlerFunc {
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
			respondError(w, errUnauthorized)
			return
		}

		accessToken := strings.TrimPrefix(authHeader, "Bearer ")
		user, err := jwtMiddleware.parseAccess(accessToken)
		if err != nil {
			respondError(w, errUnauthorized)
			return
		}

		hashBytes, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Error().Msg("auth_router.resetPassword: error generating hash:" + err.Error())
			respondInternalError(w)
			return
		}
		hashString := string(hashBytes)

		err = func() error {
			if user.Role == ResidentRole {
				return residentRepo.SetPasswordFor(user.Id, hashString)
			} else if user.Role == AdminRole {
				return adminRepo.SetPasswordFor(user.Id, hashString)
			}
			return nil
		}()
		if err != nil {
			log.Error().Msg("auth_router.resetPassword: " + err.Error())
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, message{"Password has been successfully reset."})
	}
}

// helpers
func getUserAndHashById(id string, adminRepo storage.AdminRepo, residentRepo storage.ResidentRepo) (user, string, error) {
	var userFound user
	var hash string

	if !regexp.MustCompile("^(B|T)\\d{7}$").MatchString(id) {
		admin, err := adminRepo.GetOne(id)
		if errors.Is(err, storage.ErrNoRows) {
			return user{}, "", errUnauthorized
		} else if err != nil {
			return user{}, "", fmt.Errorf("Error querying adminRepo: %v", err)
		}

		userFound = newUser(admin.Id, admin.FirstName, admin.LastName, admin.Email, AdminRole, admin.TokenVersion)
		hash = admin.Password
	} else {
		resident, err := residentRepo.GetOne(id)
		if errors.Is(err, storage.ErrNoRows) {
			return user{}, "", errUnauthorized
		} else if err != nil {
			return user{}, "", fmt.Errorf("Error querying residentRepo: %v", err)
		}

		userFound = newUser(resident.Id, resident.FirstName, resident.LastName, resident.Email, ResidentRole, resident.TokenVersion)
		hash = resident.Password
	}

	return userFound, hash, nil
}

func sendRefreshToken(w http.ResponseWriter, refreshToken string) {
	cookie := http.Cookie{
		Name:     refreshCookieKey,
		Value:    refreshToken,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode}
	http.SetCookie(w, &cookie)
}

func getGmailService(ctx context.Context, oauthConfig config.OAuthConfig) (*gmail.Service, error) {
	config := &oauth2.Config{
		ClientID:     oauthConfig.ClientID(),
		ClientSecret: oauthConfig.ClientSecret(),
		RedirectURL:  oauthConfig.RedirectURL(),
		Scopes:       []string{oauthConfig.Scope()},
		Endpoint: oauth2.Endpoint{
			AuthURL:  oauthConfig.AuthURL(),
			TokenURL: oauthConfig.TokenURL(),
		},
	}

	token := &oauth2.Token{
		AccessToken:  oauthConfig.AccessToken(),
		RefreshToken: oauthConfig.RefreshToken(),
		TokenType:    oauthConfig.TokenType(),
		Expiry:       oauthConfig.Expiry(),
	}

	client := config.Client(ctx, token)

	service, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve Gmail client: %v", err)
	}

	return service, nil
}

func createGmailMessage(jwtMiddleware jwtMiddleware, toUser user) (*gmail.Message, error) {
	body := &bytes.Buffer{}

	token, err := jwtMiddleware.newAccess(toUser.Id, toUser.Role)
	if err != nil {
		return nil, fmt.Errorf("Error generating JWT: %v", err)
	}

	fmt.Fprintf(body, "From: Park Spot <parkspotapplication@gmail.com>\r\n")
	fmt.Fprintf(body, "To: %s %s <%s>\r\n", toUser.FirstName, toUser.LastName, toUser.Email)
	fmt.Fprintf(body, "Subject: Password Reset\r\n")
	fmt.Fprintf(body, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(body, "Content-Type: text/html\r\n")
	fmt.Fprintf(body, `
    <body style='text-align: center;'>
        <h1>Password Reset for Account %s</h1>
        <p>Hi, a password reset was requested.</p>
        <p>If you sent the request, please click the button below to reset your password.
           Otherwise, you can ignore this email.</p>
        <a href='parkspotapp.com/reset-password?token=%s'>Reset Your Password</a>
    </body>`, toUser.Id, token)

	gmailMessage := &gmail.Message{Raw: base64.URLEncoding.EncodeToString(body.Bytes())}

	return gmailMessage, nil
}
