package config

import (
	"fmt"
	"os"
	"time"
)

type OAuthConfig struct {
	clientID     string
	clientSecret string
	redirectURL  string
	scope        string
	authURL      string
	tokenURL     string
	accessToken  string
	refreshToken string
	tokenType    string
	expiry       time.Time
}

func newOAuthConfig() (OAuthConfig, error) {
	config := OAuthConfig{}

	if config.clientID = os.Getenv("OAUTH_CLIENTID"); config.clientID == "" {
		return OAuthConfig{}, fmt.Errorf("OAUTH_CLIENTID is required.")
	} else if config.clientSecret = os.Getenv("OAUTH_CLIENTSECRET"); config.clientSecret == "" {
		return OAuthConfig{}, fmt.Errorf("OAUTH_CLIENTSECRET is required.")
	} else if config.redirectURL = os.Getenv("OAUTH_REDIRECTURL"); config.redirectURL == "" {
		return OAuthConfig{}, fmt.Errorf("OAUTH_REDIRECTURL is required.")
	} else if config.scope = os.Getenv("OAUTH_SCOPE"); config.scope == "" {
		return OAuthConfig{}, fmt.Errorf("OAUTH_SCOPE is required.")
	} else if config.authURL = os.Getenv("OAUTH_AUTHURL"); config.authURL == "" {
		return OAuthConfig{}, fmt.Errorf("OAUTH_AUTHURL is required.")
	} else if config.tokenURL = os.Getenv("OAUTH_TOKENURL"); config.tokenURL == "" {
		return OAuthConfig{}, fmt.Errorf("OAUTH_TOKENURL is required.")
	} else if config.accessToken = os.Getenv("OAUTH_ACCESSTOKEN"); config.accessToken == "" {
		return OAuthConfig{}, fmt.Errorf("OAUTH_ACCESSTOKEN is required.")
	} else if config.refreshToken = os.Getenv("OAUTH_REFRESHTOKEN"); config.refreshToken == "" {
		return OAuthConfig{}, fmt.Errorf("OAUTH_REFRESHTOKEN is required.")
	} else if config.tokenType = os.Getenv("OAUTH_TOKENTYPE"); config.tokenType == "" {
		return OAuthConfig{}, fmt.Errorf("OAUTH_TOKENTYPE is required.")
	} else if expiryStr := os.Getenv("OAUTH_EXPIRY"); expiryStr == "" {
		return OAuthConfig{}, fmt.Errorf("OAUTH_EXPIRY is required.")
	} else if expiryAsTime, err := time.Parse("2006-01-02T15:04:05.000000-07:00", expiryStr); err != nil {
		return OAuthConfig{}, fmt.Errorf("OAUTH_EXPIRY malformed.")
	} else {
		config.expiry = expiryAsTime
	}

	return config, nil
}

func (oauthConfig OAuthConfig) ClientID() string {
	return oauthConfig.clientID
}
func (oauthConfig OAuthConfig) ClientSecret() string {
	return oauthConfig.clientSecret
}
func (oauthConfig OAuthConfig) RedirectURL() string {
	return oauthConfig.redirectURL
}
func (oauthConfig OAuthConfig) Scope() string {
	return oauthConfig.scope
}
func (oauthConfig OAuthConfig) AuthURL() string {
	return oauthConfig.authURL
}
func (oauthConfig OAuthConfig) TokenURL() string {
	return oauthConfig.tokenURL
}
func (oauthConfig OAuthConfig) AccessToken() string {
	return oauthConfig.accessToken
}
func (oauthConfig OAuthConfig) RefreshToken() string {
	return oauthConfig.refreshToken
}
func (oauthConfig OAuthConfig) TokenType() string {
	return oauthConfig.tokenType
}
func (oauthConfig OAuthConfig) Expiry() time.Time {
	return oauthConfig.expiry
}
