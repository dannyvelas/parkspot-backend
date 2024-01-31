package email

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var _ Sender = GmailService{}

type GmailService struct {
	gmailService *gmail.Service
}

func NewGmailService(ctx context.Context, oauthConfig config.OAuthConfig) (GmailService, error) {
	config := &oauth2.Config{
		ClientID:     oauthConfig.ClientID,
		ClientSecret: oauthConfig.ClientSecret,
		RedirectURL:  oauthConfig.RedirectURL,
		Scopes:       []string{oauthConfig.Scope},
		Endpoint: oauth2.Endpoint{
			AuthURL:  oauthConfig.AuthURL,
			TokenURL: oauthConfig.TokenURL,
		},
	}

	token := &oauth2.Token{
		AccessToken:  oauthConfig.AccessToken,
		RefreshToken: oauthConfig.RefreshToken,
		TokenType:    oauthConfig.TokenType,
		Expiry:       oauthConfig.Expiry,
	}

	client := config.Client(ctx, token)

	service, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return GmailService{}, fmt.Errorf("Unable to retrieve Gmail client: %v", err)
	}

	return GmailService{service}, nil
}

func (s GmailService) Send(body []byte) error {
	gmailMessage := &gmail.Message{Raw: base64.URLEncoding.EncodeToString(body)}
	_, err := s.gmailService.Users.Messages.Send("me", gmailMessage).Do()
	return err
}
