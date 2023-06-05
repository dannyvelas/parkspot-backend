package app

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type AuthService struct {
	jwtService      JWTService
	adminService    AdminService
	residentService ResidentService
	httpConfig      config.HttpConfig
	oauthConfig     config.OAuthConfig
}

func NewAuthService(
	jwtService JWTService,
	adminService AdminService,
	residentService ResidentService,
	httpConfig config.HttpConfig,
	oauthConfig config.OAuthConfig,
) AuthService {
	return AuthService{
		jwtService:      jwtService,
		adminService:    adminService,
		residentService: residentService,
		httpConfig:      httpConfig,
		oauthConfig:     oauthConfig,
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
		log.Info().Msgf("user not found")
		log.Debug().Msgf("user not found")
		return Session{}, "", errs.Unauthorized
	} else if err != nil {
		return Session{}, "", fmt.Errorf("auth_service.refreshTokens: error querying repo: %v", err)
	}

	userFromDB := loginable.AsUser()
	if userFromDB.TokenVersion != user.TokenVersion {
		log.Info().Msgf("token version not same. %s != %s", userFromDB.TokenVersion, user.TokenVersion)
		log.Debug().Msgf("token version not same. %s != %s", userFromDB.TokenVersion, user.TokenVersion)
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

	service, err := a.getGmailService(ctx)
	if err != nil {
		return fmt.Errorf("auth_service.sendResetPasswordEmail: %v", err)
	}

	gmailMessage, err := a.createGmailMessage(loginable.AsUser())
	if err != nil {
		return fmt.Errorf("auth_service.sendResetPasswordEmail: %v", err)
	}

	_, err = service.Users.Messages.Send("me", gmailMessage).Do()
	if err != nil {
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
		_, err = a.adminService.Update(models.Admin{Password: newPass})
	} else {
		_, err = a.residentService.Update(models.Resident{Password: newPass})
	}

	if err != nil {
		return fmt.Errorf("authService.resetPassword: error updating password: %v", err)
	}

	return nil
}

func (a AuthService) getGmailService(ctx context.Context) (*gmail.Service, error) {
	config := &oauth2.Config{
		ClientID:     a.oauthConfig.ClientID,
		ClientSecret: a.oauthConfig.ClientSecret,
		RedirectURL:  a.oauthConfig.RedirectURL,
		Scopes:       []string{a.oauthConfig.Scope},
		Endpoint: oauth2.Endpoint{
			AuthURL:  a.oauthConfig.AuthURL,
			TokenURL: a.oauthConfig.TokenURL,
		},
	}

	token := &oauth2.Token{
		AccessToken:  a.oauthConfig.AccessToken,
		RefreshToken: a.oauthConfig.RefreshToken,
		TokenType:    a.oauthConfig.TokenType,
		Expiry:       a.oauthConfig.Expiry,
	}

	client := config.Client(ctx, token)

	service, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve Gmail client: %v", err)
	}

	return service, nil
}

func (a AuthService) createGmailMessage(toUser models.User) (*gmail.Message, error) {
	body := &bytes.Buffer{}

	token, err := a.jwtService.NewAccess(toUser.ID, toUser.Role)
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
    </body>`, a.httpConfig.Domain, token)

	gmailMessage := &gmail.Message{Raw: base64.URLEncoding.EncodeToString(body.Bytes())}

	return gmailMessage, nil
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
