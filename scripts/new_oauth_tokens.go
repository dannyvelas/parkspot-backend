package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/dannyvelas/parkspot-backend/config"
	"golang.org/x/oauth2"
)

func main() {
	// load config
	c, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// create oauth config service
	config := &oauth2.Config{
		ClientID:     c.OAuth.ClientID,
		ClientSecret: c.OAuth.ClientSecret,
		RedirectURL:  c.OAuth.RedirectURL,
		Scopes:       []string{c.OAuth.Scope},
		Endpoint: oauth2.Endpoint{
			AuthURL:  c.OAuth.AuthURL,
			TokenURL: c.OAuth.TokenURL,
		},
	}

	// generate verifier
	verifier := oauth2.GenerateVerifier()

	url := config.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
	fmt.Printf("Visit the URL for the auth dialog: %v\n", url)
	fmt.Printf("Once you are done, you should be redirected to: %v. The URL will have an auth code as a query parameter. copy and paste it here:\n", c.OAuth.RedirectURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("error scanning: %v", err)
	}

	tokens, err := config.Exchange(context.Background(), code, oauth2.VerifierOption(verifier))
	if err != nil {
		log.Fatal(err)
	}

	bytes, err := json.Marshal(tokens)
	if err != nil {
		log.Fatalf("failure to marshal tokens %v. error was: %v", tokens, err)
	}

	fmt.Println(string(bytes))
}
