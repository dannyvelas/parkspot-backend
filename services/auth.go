package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type AuthService struct {
	jwtService   JWTService
	adminRepo    storage.AdminRepo
	residentRepo storage.ResidentRepo
	httpConfig   config.HttpConfig
	oauthConfig  config.OAuthConfig
}

func NewAuthService(
	jwtService JWTService,
	adminRepo storage.AdminRepo,
	residentRepo storage.ResidentRepo,
	httpConfig config.HttpConfig,
	oauthConfig config.OAuthConfig,
) AuthService {
	return AuthService{
		jwtService:   jwtService,
		adminRepo:    adminRepo,
		residentRepo: residentRepo,
		httpConfig:   httpConfig,
		oauthConfig:  oauthConfig,
	}
}

type Session struct {
	User        models.User `json:"user"`
	AccessToken string      `json:"accessToken"`
}

func (a AuthService) Login(id, password string) (Session, string, error) {
	loginable, err := a.getUser(id)
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
	refreshToken, err := a.jwtService.newRefresh(user)
	if err != nil {
		return Session{}, "", fmt.Errorf("auth_service.login: Error generating refresh JWT: %v", err)
	}

	accessToken, err := a.jwtService.newAccess(user.Id, user.Role)
	if err != nil {
		return Session{}, "", fmt.Errorf("auth_service.login: Error generating access JWT: %v", err)
	}

	return Session{user, accessToken}, refreshToken, nil
}

func (a AuthService) RefreshTokens(user models.User) (Session, string, error) {
	loginable, err := a.getUser(user.Id)
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
	refreshToken, err := a.jwtService.newRefresh(user)
	if err != nil {
		return Session{}, "", fmt.Errorf("auth_service.refreshTokens: Error generating refresh JWT: %v", err)
	}

	accessToken, err := a.jwtService.newAccess(user.Id, user.Role)
	if err != nil {
		return Session{}, "", fmt.Errorf("auth_service.refreshTokens: Error generating access JWT: %v", err)
	}

	return Session{user, accessToken}, refreshToken, nil
}

func (a AuthService) SendResetPasswordEmail(ctx context.Context, id string) error {
	loginable, err := a.getUser(id)
	if errors.Is(err, storage.ErrNoRows) {
		return ErrUnauthorized
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
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("auth_router.resetPassword: error generating hash: %v", err)
	}

	var userRepo interface {
		SetPassword(id, password string) error
	}
	if resCheckErr := models.IsResidentId(id); resCheckErr != nil {
		userRepo = a.adminRepo
	} else {
		userRepo = a.residentRepo
	}

	err = userRepo.SetPassword(id, string(hashBytes))
	if err != nil {
		return fmt.Errorf("auth_router.resetPassword: error updating password: %v", err)
	}

	return nil
}

func (a AuthService) getGmailService(ctx context.Context) (*gmail.Service, error) {
	config := &oauth2.Config{
		ClientID:     a.oauthConfig.ClientID(),
		ClientSecret: a.oauthConfig.ClientSecret(),
		RedirectURL:  a.oauthConfig.RedirectURL(),
		Scopes:       []string{a.oauthConfig.Scope()},
		Endpoint: oauth2.Endpoint{
			AuthURL:  a.oauthConfig.AuthURL(),
			TokenURL: a.oauthConfig.TokenURL(),
		},
	}

	token := &oauth2.Token{
		AccessToken:  a.oauthConfig.AccessToken(),
		RefreshToken: a.oauthConfig.RefreshToken(),
		TokenType:    a.oauthConfig.TokenType(),
		Expiry:       a.oauthConfig.Expiry(),
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

	token, err := a.jwtService.newAccess(toUser.Id, toUser.Role)
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
    </body>`, a.httpConfig.Domain(), token)

	gmailMessage := &gmail.Message{Raw: base64.URLEncoding.EncodeToString(body.Bytes())}

	return gmailMessage, nil
}

func (a AuthService) getUser(id string) (loginable models.Loginable, err error) {
	// Unfortunately, dynamic dispatch via a common iface is impossible as both `GetOne` fns have diff return types
	if resCheckErr := models.IsResidentId(id); resCheckErr != nil {
		loginable, err = a.adminRepo.GetOne(id)
	} else {
		loginable, err = a.residentRepo.GetOne(id)
	}
	return
}
