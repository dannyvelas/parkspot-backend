package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/email"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	jwtService      JWTService
	adminService    AdminService
	residentService ResidentService
	emailSender     email.Sender
	httpConfig      config.HttpConfig
}

func NewAuthService(
	jwtService JWTService,
	adminService AdminService,
	residentService ResidentService,
	emailSender email.Sender,
	httpConfig config.HttpConfig,
) AuthService {
	return AuthService{
		jwtService:      jwtService,
		adminService:    adminService,
		residentService: residentService,
		emailSender:     emailSender,
		httpConfig:      httpConfig,
	}
}

type Session struct {
	User        models.User `json:"user"`
	AccessToken string      `json:"accessToken"`
}

func (a AuthService) Login(id, password string) (Session, string, error) {
	loginable, err := a.getUser(id)
	if errors.Is(err, errs.NotFound) {
		return Session{}, "", errs.Unauthorized
	} else if err != nil {
		return Session{}, "", fmt.Errorf("auth_service.login: error querying repo: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(loginable.GetPassword()),
		[]byte(password),
	); err != nil {
		return Session{}, "", errs.Unauthorized
	}

	user := loginable.AsUser()

	// generate tokens
	refreshToken, err := a.jwtService.NewRefresh(user)
	if err != nil {
		return Session{}, "", fmt.Errorf("auth_service.login: Error generating refresh JWT: %v", err)
	}

	accessToken, err := a.jwtService.NewAccess(user.ID, user.Role)
	if err != nil {
		return Session{}, "", fmt.Errorf("auth_service.login: Error generating access JWT: %v", err)
	}

	return Session{user, accessToken}, refreshToken, nil
}

func (a AuthService) RefreshTokens(user models.User) (Session, string, error) {
	loginable, err := a.getUser(user.ID)
	if errors.Is(err, errs.NotFound) {
		return Session{}, "", errs.Unauthorized
	} else if err != nil {
		return Session{}, "", fmt.Errorf("auth_service.refreshTokens: error querying repo: %v", err)
	}

	userFromDB := loginable.AsUser()
	if userFromDB.TokenVersion != user.TokenVersion {
		log.Debug().
			Int("dbTokenVersion", userFromDB.TokenVersion).
			Int("tokenVersion", user.TokenVersion).
			Msgf("token version not same")
		return Session{}, "", errs.Unauthorized
	}

	// generate tokens
	refreshToken, err := a.jwtService.NewRefresh(user)
	if err != nil {
		return Session{}, "", fmt.Errorf("auth_service.refreshTokens: Error generating refresh JWT: %v", err)
	}

	accessToken, err := a.jwtService.NewAccess(user.ID, user.Role)
	if err != nil {
		return Session{}, "", fmt.Errorf("auth_service.refreshTokens: Error generating access JWT: %v", err)
	}

	return Session{user, accessToken}, refreshToken, nil
}

func (a AuthService) SendResetPasswordEmail(ctx context.Context, id string) error {
	loginable, err := a.getUser(id)
	if errors.Is(err, errs.NotFound) {
		return errs.Unauthorized
	} else if err != nil {
		return fmt.Errorf("auth_service.sendResetPasswordEmail: error querying repo: %v", err)
	}

	message, err := a.createMessage(loginable.AsUser())
	if err != nil {
		return fmt.Errorf("auth_service.sendResetPasswordEmail: %v", err)
	}

	if err := a.emailSender.Send(message); err != nil {
		return fmt.Errorf("auth_service.sendResetPasswordEmail: error sending mail: %v", err)
	}

	return nil
}

func (a AuthService) ResetPassword(id, newPass string) error {
	if id == "" {
		return errs.MissingIDField
	} else if newPass == "" {
		return errs.EmptyFields("password")
	}

	var err error
	if resCheckErr := models.IsResidentID(id); resCheckErr != nil {
		_, err = a.adminService.Update(models.Admin{ID: id, Password: newPass})
	} else {
		_, err = a.residentService.Update(models.Resident{ID: id, Password: newPass})
	}

	if err != nil {
		return fmt.Errorf("authService.resetPassword: error updating password: %v", err)
	}

	return nil
}

func (a AuthService) createMessage(toUser models.User) ([]byte, error) {
	token, err := a.jwtService.NewAccess(toUser.ID, toUser.Role)
	if err != nil {
		return nil, fmt.Errorf("Error generating JWT: %v", err)
	}

	body := &bytes.Buffer{}
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
    </body>`, a.httpConfig.FrontendURL, token)

	return body.Bytes(), nil
}

func (a AuthService) getUser(id string) (models.Loginable, error) {
	// i wanted to define an interface that both adminService and residentService implement:
	// type UserService interface { GetOne(id string) (models.Loginable, error) }
	// but adminService and residentService cannot implement the GetOne function
	// because they return different types (models.Admin) and (models.User).
	// and, go mandates that structs must return the same exact type to implement a common interface
	if resCheckErr := models.IsResidentID(id); resCheckErr != nil {
		return a.adminService.GetOne(id)
	} else {
		return a.residentService.GetOne(id)
	}
}
