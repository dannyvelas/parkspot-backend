package services

import (
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceIface interface {
	Login(id, password string) (Session, string, error)
	RefreshTokens(user models.User) (Session, string, error)
}

type AuthService[T models.Loginable] struct {
	JWTService JWTService
	UserRepo   storage.UserRepo[T]
}

func NewAuthService[T models.Loginable](jwtService JWTService, userRepo storage.UserRepo[T]) AuthService[T] {
	return AuthService[T]{
		JWTService: jwtService,
		UserRepo:   userRepo,
	}
}

type Session struct {
	User        models.User `json:"user"`
	AccessToken string      `json:"accessToken"`
}

func (a AuthService[T]) Login(id, password string) (Session, string, error) {
	loginable, err := a.UserRepo.GetOne(id)
	if errors.Is(err, storage.ErrNoRows) {
		return Session{}, "", ErrUnauthorized
	} else if err != nil {
		return Session{}, "", fmt.Errorf("auth_service.login: error querying repo: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(loginable.GetPassword()),
		[]byte(password),
	); err != nil {
		return Session{}, "", ErrUnauthorized
	}

	user := loginable.AsUser()

	// generate tokens
	refreshToken, err := a.JWTService.newRefresh(user)
	if err != nil {
		return Session{}, "", fmt.Errorf("auth_service.login: Error generating refresh JWT: %v", err)
	}

	accessToken, err := a.JWTService.newAccess(user.Id, user.Role)
	if err != nil {
		return Session{}, "", fmt.Errorf("auth_service.login: Error generating access JWT: %v", err)
	}

	return Session{user, accessToken}, refreshToken, nil
}

func (a AuthService[T]) RefreshTokens(user models.User) (Session, string, error) {
	loginable, err := a.UserRepo.GetOne(user.Id)
	if errors.Is(err, storage.ErrNoRows) {
		return Session{}, "", ErrUnauthorized
	} else if err != nil {
		return Session{}, "", fmt.Errorf("auth_service.refreshTokens: error querying repo: %v", err)
	}

	userFromDB := loginable.AsUser()

	if userFromDB.TokenVersion != user.TokenVersion {
		return Session{}, "", ErrUnauthorized
	}

	// generate tokens
	refreshToken, err := a.JWTService.newRefresh(user)
	if err != nil {
		return Session{}, "", fmt.Errorf("auth_service.refreshTokens: Error generating refresh JWT: %v", err)
	}

	accessToken, err := a.JWTService.newAccess(user.Id, user.Role)
	if err != nil {
		return Session{}, "", fmt.Errorf("auth_service.refreshTokens: Error generating access JWT: %v", err)
	}

	return Session{user, accessToken}, refreshToken, nil
}

func sendResetPasswordEmail(
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
