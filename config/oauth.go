package config

import (
	"fmt"
	"os"
	"time"
)

type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scope        string
	AuthURL      string
	TokenURL     string
	AccessToken  string
	RefreshToken string
	TokenType    string
	Expiry       time.Time
}

func newOAuthConfig() (OAuthConfig, error) {
	config := OAuthConfig{}

	if config.ClientID = os.Getenv("OAUTH_CLIENTID"); config.ClientID == "" {
		return OAuthConfig{}, fmt.Errorf("OAUTH_CLIENTID is required.")
	} else if config.ClientSecret = os.Getenv("OAUTH_CLIENTSECRET"); config.ClientSecret == "" {
		return OAuthConfig{}, fmt.Errorf("OAUTH_CLIENTSECRET is required.")
	} else if config.RedirectURL = os.Getenv("OAUTH_REDIRECTURL"); config.RedirectURL == "" {
		return OAuthConfig{}, fmt.Errorf("OAUTH_REDIRECTURL is required.")
	} else if config.Scope = os.Getenv("OAUTH_SCOPE"); config.Scope == "" {
		return OAuthConfig{}, fmt.Errorf("OAUTH_SCOPE is required.")
	} else if config.AuthURL = os.Getenv("OAUTH_AUTHURL"); config.AuthURL == "" {
		return OAuthConfig{}, fmt.Errorf("OAUTH_AUTHURL is required.")
	} else if config.TokenURL = os.Getenv("OAUTH_TOKENURL"); config.TokenURL == "" {
		return OAuthConfig{}, fmt.Errorf("OAUTH_TOKENURL is required.")
	} else if config.AccessToken = os.Getenv("OAUTH_ACCESSTOKEN"); config.AccessToken == "" {
		return OAuthConfig{}, fmt.Errorf("OAUTH_ACCESSTOKEN is required.")
	} else if config.RefreshToken = os.Getenv("OAUTH_REFRESHTOKEN"); config.RefreshToken == "" {
		return OAuthConfig{}, fmt.Errorf("OAUTH_REFRESHTOKEN is required.")
	} else if config.TokenType = os.Getenv("OAUTH_TOKENTYPE"); config.TokenType == "" {
		return OAuthConfig{}, fmt.Errorf("OAUTH_TOKENTYPE is required.")
	} else if expiryStr := os.Getenv("OAUTH_EXPIRY"); expiryStr == "" {
		return OAuthConfig{}, fmt.Errorf("OAUTH_EXPIRY is required.")
	} else if expiryAsTime, err := time.Parse("2006-01-02T15:04:05.000000-07:00", expiryStr); err != nil {
		return OAuthConfig{}, fmt.Errorf("OAUTH_EXPIRY malformed.")
	} else {
		config.Expiry = expiryAsTime
	}

	return config, nil
}
