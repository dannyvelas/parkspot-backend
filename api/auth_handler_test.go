package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/app"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/util"
	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type authRouterSuite struct {
	suite.Suite
	testServer        *httptest.Server
	app               app.App
	adminUser         models.User
	adminUserPassword string
	adminAccessToken  string
}

func TestAuthRouter(t *testing.T) {
	suite.Run(t, new(authRouterSuite))
}

func (suite *authRouterSuite) SetupSuite() {
	c, err := config.NewConfig()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	suite.app, err = app.NewApp(c)
	if err != nil {
		log.Fatal().Msgf("Failed to initialize app: %v", err)
	}

	router := newRouter(c, suite.app)
	suite.testServer = httptest.NewServer(router)

	suite.adminUser = models.NewUser("test",
		"Daniel",
		"Velasquez",
		"email@example.com",
		models.AdminRole,
		0)
	suite.adminUserPassword = "notapassword"

	suite.adminAccessToken, err = suite.app.JWTService.NewAccess(suite.adminUser.ID, models.AdminRole)
	if err != nil {
		log.Fatal().Msgf("Failed to create JWT: %v", err)
	}

	if err := suite.app.ResidentService.Create(testResident); err != nil {
		log.Fatal().Msgf("error creating test resident: %v", err.Error())
	}
}

func (suite authRouterSuite) TearDownSuite() {
	defer suite.testServer.Close()

	if err := suite.app.ResidentService.Delete(testResident.ID); err != nil {
		log.Fatal().Msgf("error deleting test resident: %v", err.Error())
	}
}

func (suite authRouterSuite) TestLogin_Admin_Positive() {
	requestBody := []byte(fmt.Sprintf(`{
    "id":"%s",
    "password":"%s"
  }`,
		suite.adminUser.ID,
		suite.adminUserPassword))
	request, err := http.NewRequest("POST", suite.testServer.URL+"/api/login", bytes.NewBuffer(requestBody))
	if err != nil {
		suite.NoError(err)
		return
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		suite.NoError(err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			suite.NoError(err)
			return
		}
		suite.Empty(string(bodyBytes))
		return
	}

	var session app.Session
	if err := json.NewDecoder(response.Body).Decode(&session); err != nil {
		suite.NoError(err)
		return
	}

	suite.Empty(cmp.Diff(suite.adminUser, session.User), "user in response did not equal expected user")

	err = checkAccessToken(suite.app.JWTService, session.AccessToken, suite.adminUser)
	suite.NoError(err)

	err = checkRefreshToken(suite.app.JWTService, response.Cookies(), suite.adminUser.ID, suite.adminUser.TokenVersion)
	suite.NoError(err)
}

func (suite authRouterSuite) TestLogin_Resident_Positive() {
	requestBody := []byte(fmt.Sprintf(`{
    "id":"%s",
    "password":"%s"
  }`, testResident.ID, testResident.Password))
	request, err := http.NewRequest("POST", suite.testServer.URL+"/api/login", bytes.NewBuffer(requestBody))
	if err != nil {
		suite.NoError(err)
		return
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		suite.NoError(err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			suite.NoError(err)
			return
		}
		suite.Empty(string(bodyBytes))
		return
	}

	var session app.Session
	if err := json.NewDecoder(response.Body).Decode(&session); err != nil {
		suite.NoError(err)
		return
	}

	expectedUser := models.NewUser(testResident.ID,
		testResident.FirstName,
		testResident.LastName,
		testResident.Email,
		models.ResidentRole,
		0)
	suite.Empty(cmp.Diff(expectedUser, session.User), "user in response did not equal expected user")

	err = checkAccessToken(suite.app.JWTService, session.AccessToken, expectedUser)
	suite.NoError(err)

	err = checkRefreshToken(suite.app.JWTService, response.Cookies(), expectedUser.ID, expectedUser.TokenVersion)
	suite.NoError(err)
}

func (suite authRouterSuite) TestRefreshTokens_Positive() {
	request, err := http.NewRequest("POST", suite.testServer.URL+"/api/refresh-tokens", bytes.NewBuffer([]byte{}))
	if err != nil {
		suite.NoError(fmt.Errorf("error creating http.Request: %s", err))
		return
	}

	user := models.NewUser(testResident.ID,
		testResident.FirstName,
		testResident.LastName,
		testResident.Email,
		models.ResidentRole,
		*testResident.TokenVersion)
	refreshToken, err := suite.app.JWTService.NewRefresh(user)
	if err != nil {
		suite.NoError(fmt.Errorf("error creating refresh token: %s", err))
		return
	}
	cookie := http.Cookie{Name: config.RefreshCookieKey, Value: refreshToken, HttpOnly: true, Path: "/"}
	request.AddCookie(&cookie)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		suite.NoError(fmt.Errorf("error sending http.Request: %s", err))
		return
	}

	if response.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			suite.NoError(err)
			return
		}
		suite.Empty("Unexpected bad response: " + string(bodyBytes))
		return
	}

	var session app.Session
	if err := json.NewDecoder(response.Body).Decode(&session); err != nil {
		suite.NoError(err)
		return
	}

	expectedUser := models.NewUser(testResident.ID,
		testResident.FirstName,
		testResident.LastName,
		testResident.Email,
		models.ResidentRole,
		0)
	suite.Empty(cmp.Diff(expectedUser, session.User), "user in response did not equal expected user")

	err = checkAccessToken(suite.app.JWTService, session.AccessToken, expectedUser)
	suite.NoError(err)

	err = checkRefreshToken(suite.app.JWTService, response.Cookies(), expectedUser.ID, expectedUser.TokenVersion)
	suite.NoError(err)
}

// helpers
func checkAccessToken(jwtService app.JWTService, token string, expectedUser models.User) error {
	if token == "" {
		return fmt.Errorf("accessToken was empty")
	} else if payload, err := jwtService.ParseAccess(token); err != nil {
		return fmt.Errorf("Error parsing access token (%s): %v", token, err)
	} else if expectedUser.ID != payload.ID {
		return fmt.Errorf("user id (%s) was not the same to access payload id (%s)", expectedUser.ID, payload.ID)
	} else if expectedUser.Role != payload.Role {
		return fmt.Errorf("user role (%v) was not the same to access payload role (%v)", expectedUser.Role, payload.Role)
	}

	return nil
}

func checkRefreshToken(jwtService app.JWTService, cookies []*http.Cookie, expectedID string, expectedVersion int) error {
	index := util.Find(cookies, func(cookie *http.Cookie) bool {
		return cookie.Name == config.RefreshCookieKey
	})
	if index == -1 {
		return fmt.Errorf("cookie with key of %s not found", config.RefreshCookieKey)
	}
	refreshCookie := cookies[index]

	if payload, err := jwtService.ParseRefresh(refreshCookie.Value); err != nil {
		return fmt.Errorf("Error parsing refresh token (%s): %v", refreshCookie.Value, err)
	} else if expectedID != payload.ID {
		return fmt.Errorf("user id (%s) was not the same to refresh payload id (%s)", expectedID, payload.ID)
	} else if expectedVersion != payload.TokenVersion {
		return fmt.Errorf("user version (%v) was not the same to refresh payload version (%v)", expectedVersion, payload.TokenVersion)
	}

	return nil
}
