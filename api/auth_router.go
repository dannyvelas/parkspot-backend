package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"
)

type credentials struct {
	Id       string
	Password string
}

func login(jwtMiddleware jwtMiddleware, adminRepo storage.AdminRepo, residentRepo storage.ResidentRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds credentials
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			respondError(w, newErrMalformed("Credentials"))
			return
		}

		user, hash, err := getUserAndHashById(creds.Id, adminRepo, residentRepo)
		if errors.Is(err, errUnauthorized) {
			respondError(w, errUnauthorized)
			return
		} else if err != nil {
			log.Error().Msg("auth_router: " + err.Error())
			respondInternalError(w)
			return
		}

		if err := bcrypt.CompareHashAndPassword(
			[]byte(hash),
			[]byte(creds.Password),
		); err != nil {
			respondError(w, errUnauthorized)
			return
		}

		token, err := jwtMiddleware.newJWT(user.Id, user.FirstName, user.LastName, user.Email, user.Role)
		if err != nil {
			log.Error().Msgf("auth_router: Error generating JWT: %v", err)
			respondInternalError(w)
			return
		}

		cookie := http.Cookie{Name: "jwt", Value: token, HttpOnly: true, Path: "/"}
		http.SetCookie(w, &cookie)

		respondJSON(w, http.StatusOK, user)
	}
}

func logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie := http.Cookie{Name: "jwt", Value: "deleted", HttpOnly: true, Path: "/", Expires: time.Unix(0, 0)}
		http.SetCookie(w, &cookie)

		respondJSON(w, http.StatusOK, message{"Successfully logged-out user"})
	}
}

func sendResetPasswordEmail(jwtMiddleware jwtMiddleware, adminRepo storage.AdminRepo, residentRepo storage.ResidentRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var idStruct struct{ Id string }
		if err := json.NewDecoder(r.Body).Decode(&idStruct); err != nil {
			respondError(w, newErrMalformed("id object"))
			return
		}

		userFound, _, err := getUserAndHashById(idStruct.Id, adminRepo, residentRepo)
		if errors.Is(err, errUnauthorized) {
			respondError(w, errUnauthorized)
			return
		} else if err != nil {
			log.Error().Msg("auth_router.sendResetPasswordEmail: " + err.Error())
			respondInternalError(w)
			return
		}

		service, err := getGmailService(r.Context())
		if err != nil {
			log.Error().Msgf(err.Error())
			respondInternalError(w)
			return
		}

		gmailMessage, err := createGmailMessage(jwtMiddleware, userFound)
		if err != nil {
			log.Error().Msgf(err.Error())
			respondInternalError(w)
			return
		}

		_, err = service.Users.Messages.Send("me", &gmailMessage).Do()
		if err != nil {
			log.Error().Msg("auth_router.sendResetPasswordEmail: error sending mail:" + err.Error())
			respondInternalError(w)
			return
		}

		// send email
		respondJSON(w, http.StatusOK, message{"ok"})
	}
}

// helpers
func getUserAndHashById(username string, adminRepo storage.AdminRepo, residentRepo storage.ResidentRepo) (user, string, error) {
	var userFound user
	var hash string

	if !regexp.MustCompile("^(B|T)\\d{7}$").MatchString(username) {
		admin, err := adminRepo.GetOne(username)
		if errors.Is(err, storage.ErrNoRows) {
			return user{}, "", errUnauthorized
		} else if err != nil {
			return user{}, "", fmt.Errorf("Error querying adminRepo: %v", err)
		}

		userFound = newUser(admin.Id, admin.FirstName, admin.LastName, admin.Email, AdminRole)
		hash = admin.Password
	} else {
		resident, err := residentRepo.GetOne(username)
		if errors.Is(err, storage.ErrNoRows) {
			return user{}, "", errUnauthorized
		} else if err != nil {
			return user{}, "", fmt.Errorf("Error querying residentRepo: %v", err)
		}

		userFound = newUser(resident.Id, resident.FirstName, resident.LastName, resident.Email, ResidentRole)
		hash = resident.Password
	}

	return userFound, hash, nil
}

func createGmailMessage(jwtMiddleware jwtMiddleware, toUser user) (gmail.Message, error) {
	body := &bytes.Buffer{}

	token, err := jwtMiddleware.newJWT(toUser.Id, toUser.FirstName, toUser.LastName, toUser.Email, toUser.Role)
	if err != nil {
		return gmail.Message{}, fmt.Errorf("auth_router: Error generating JWT: %v", err)
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

	gmailMessage := gmail.Message{Raw: base64.URLEncoding.EncodeToString(body.Bytes())}

	return gmailMessage, nil
}

func getGmailService(ctx context.Context) (*gmail.Service, error) {
	bytes, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		return nil, fmt.Errorf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(bytes, gmail.GmailComposeScope)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse client secret file to config: %v", err)
	}

	client, err := getClient(config)
	if err != nil {
		return nil, fmt.Errorf("Unable to get client from config: %v", err)
	}

	service, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve Gmail client: %v", err)
	}

	return service, nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) (*http.Client, error) {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		log.Info().Msg("Error reading tokenFromFile")
		tok, err = getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok), nil // TODO: should i pass in the request context here?
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("Unable to read authorization code: %v", err)
	}

	token, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve token from web: %v", err)
	}
	return token, nil
}

func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)

	return nil
}
