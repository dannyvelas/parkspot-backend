package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"net/http"
	"strings"
	"time"
)

type session struct {
	User        models.User `json:"user"`
	AccessToken string      `json:"accessToken"`
}

func login[Model models.Loginable](jwtMiddleware jwtMiddleware, repo storage.UserRepo[Model]) http.HandlerFunc {
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

		loginable, err := repo.GetOne(credentials.Id)
		if errors.Is(err, storage.ErrNoRows) {
			respondError(w, errUnauthorized)
			return
		} else if err != nil {
			log.Error().Msg("auth_router.login: error querying repo: %v" + err.Error())
			respondInternalError(w)
			return
		}

		if err := bcrypt.CompareHashAndPassword(
			[]byte(loginable.GetPassword()),
			[]byte(credentials.Password),
		); err != nil {
			respondError(w, errUnauthorized)
			return
		}

		user := loginable.AsUser()

		// generate tokens
		refreshToken, err := jwtMiddleware.newRefresh(user)
		if err != nil {
			log.Error().Msgf("auth_router.login: Error generating refresh JWT: %v", err)
			respondInternalError(w)
			return
		}

		accessToken, err := jwtMiddleware.newAccess(user.Id, user.Role)
		if err != nil {
			log.Error().Msgf("auth_router.login: Error generating access JWT: %v", err)
			respondInternalError(w)
			return
		}

		// send tokens
		sendRefreshToken(w, refreshToken)
		response := session{user, accessToken}

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

		var (
			user         models.User
			refreshToken string
			accessToken  string
		)
		if resCheckErr := models.IsResidentId(refreshPayload.Id); resCheckErr == nil {
			user, refreshToken, accessToken, err = refreshService[models.Resident](jwtMiddleware, residentRepo, refreshPayload)
		} else {
			user, refreshToken, accessToken, err = refreshService[models.Admin](jwtMiddleware, adminRepo, refreshPayload)
		}
		if errors.Is(err, errUnauthorized) {
			respondError(w, errUnauthorized)
			return
		} else if err != nil {
			log.Error().Msg("auth_router.refreshTokens: " + err.Error())
			respondInternalError(w)
			return
		}

		// send tokens
		sendRefreshToken(w, refreshToken)
		response := session{user, accessToken}

		respondJSON(w, http.StatusOK, response)
	}
}

func refreshService[Model models.Loginable](
	jwtMiddleware jwtMiddleware,
	repo storage.UserRepo[Model],
	refreshPayload models.User,
) (models.User, string, string, error) {
	loginable, err := repo.GetOne(refreshPayload.Id)
	if errors.Is(err, storage.ErrNoRows) {
		return models.User{}, "", "", errUnauthorized
	} else if err != nil {
		return models.User{}, "", "", fmt.Errorf("Error querying repo: %v", err)
	}

	user := loginable.AsUser()

	if user.TokenVersion != refreshPayload.TokenVersion {
		return models.User{}, "", "", errUnauthorized
	}

	// generate tokens
	refreshToken, err := jwtMiddleware.newRefresh(user)
	if err != nil {
		return models.User{}, "", "", fmt.Errorf("auth_router.refreshService: Error generating refresh JWT: %v", err)
	}

	accessToken, err := jwtMiddleware.newAccess(user.Id, user.Role)
	if err != nil {
		return models.User{}, "", "", fmt.Errorf("auth_router.refreshService: Error generating access JWT: %v", err)
	}

	return user, refreshToken, accessToken, nil
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

		resident := models.NewResident(payload.ResidentId,
			payload.FirstName,
			payload.LastName,
			payload.Phone,
			payload.Email,
			hashString,
			payload.UnlimDays,
			0, 0)

		err = residentRepo.Create(resident)
		if err != nil {
			log.Error().Msgf("auth_router.createResident: Error querying residentRepo: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, message{"Resident successfully created."})
	}
}

func sendResetPasswordEmail(
	jwtMiddleware jwtMiddleware,
	httpConfig config.HttpConfig,
	oauthConfig config.OAuthConfig,
	adminRepo storage.AdminRepo,
	residentRepo storage.ResidentRepo,
) http.HandlerFunc {
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

		var err error
		if resCheckErr := models.IsResidentId(payload.Id); resCheckErr == nil {
			err = sendResetPasswordEmailService[models.Resident](r.Context(), jwtMiddleware, httpConfig, oauthConfig, residentRepo, payload.Id)
		} else {
			err = sendResetPasswordEmailService[models.Admin](r.Context(), jwtMiddleware, httpConfig, oauthConfig, adminRepo, payload.Id)
		}
		if errors.Is(err, errUnauthorized) {
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

func sendResetPasswordEmailService[Model models.Loginable](
	ctx context.Context,
	jwtMiddleware jwtMiddleware,
	httpConfig config.HttpConfig,
	oauthConfig config.OAuthConfig,
	repo storage.UserRepo[Model],
	id string,
) error {
	service, err := getGmailService(ctx, oauthConfig)
	if err != nil {
		return fmt.Errorf("auth_router.sendResetPasswordEmailService: %v", err)
	}

	loginable, err := repo.GetOne(id)
	if errors.Is(err, storage.ErrNoRows) {
		return errUnauthorized
	} else if err != nil {
		return fmt.Errorf("auth_router.sendResetPasswordEmailService: error querying repo: %v", err)
	}

	gmailMessage, err := createGmailMessage(jwtMiddleware, httpConfig, loginable.AsUser())
	if err != nil {
		return fmt.Errorf("auth_router.sendResetPasswordEmailService: %v", err)
	}

	_, err = service.Users.Messages.Send("me", gmailMessage).Do()
	if err != nil {
		return fmt.Errorf("auth_router.sendResetPasswordEmailService: error sending mail: %v", err)
	}

	return nil
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
			log.Debug().Msg("No 'Authorization' header was present with 'Bearer ' prefix.")
			respondError(w, errUnauthorized)
			return
		}

		accessToken := strings.TrimPrefix(authHeader, "Bearer ")
		user, err := jwtMiddleware.parseAccess(accessToken)
		if err != nil {
			log.Debug().Msgf("Error parsing: %v", err)
			respondError(w, errUnauthorized)
			return
		}

		if resCheckErr := models.IsResidentId(user.Id); resCheckErr == nil {
			err = resetPasswordService[models.Resident](jwtMiddleware, residentRepo, user.Id, payload.Password)
		} else {
			err = resetPasswordService[models.Admin](jwtMiddleware, adminRepo, user.Id, payload.Password)
		}
		if err != nil {
			log.Error().Msgf("auth_router.resetPassword: error calling service: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, message{"Password has been successfully reset."})
	}
}

func resetPasswordService[Model models.Loginable](
	jwtMiddleware jwtMiddleware,
	repo storage.UserRepo[Model],
	id string,
	newPass string,
) error {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("auth_router.resetPassword: error generating hash: %v", err)
	}

	err = repo.SetPassword(id, string(hashBytes))
	if err != nil {
		return fmt.Errorf("auth_router.resetPassword: error updating password: %v", err)
	}

	return nil
}

func sendRefreshToken(w http.ResponseWriter, refreshToken string) {
	cookie := http.Cookie{
		Name:     refreshCookieKey,
		Value:    refreshToken,
		HttpOnly: true,
		Path:     "/"}
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

func createGmailMessage(jwtMiddleware jwtMiddleware, httpConfig config.HttpConfig, toUser models.User) (*gmail.Message, error) {
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
        <h1>Password Reset</h1>
        <p>Hi, a password reset was requested.</p>
        <p>If you sent the request, please click the button below to reset your password.
           Otherwise, you can ignore this email.</p>
        <a href='%s/reset-password?token=%s'>Reset Your Password</a>
    </body>`, httpConfig.Domain(), token)

	gmailMessage := &gmail.Message{Raw: base64.URLEncoding.EncodeToString(body.Bytes())}

	return gmailMessage, nil
}
