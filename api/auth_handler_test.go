package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/app"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage/psql"
	"github.com/dannyvelas/lasvistas_api/util"
	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type authRouterSuite struct {
	suite.Suite
	container  testcontainers.Container
	testServer *httptest.Server
	app        app.App
}

func TestAuthRouter(t *testing.T) {
	suite.Run(t, new(authRouterSuite))
}

func (suite *authRouterSuite) SetupSuite() {
	c, err := config.NewConfig()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	// configure and start container
	container, database, err := psql.NewSandboxDatabase()
	if err != nil {
		suite.T().Fatalf("error getting sandbox database: %v", err)
	}
	// save container in suite struct so we can terminate it on suite teardown
	suite.container = container

	suite.app = app.NewApp(c, database)

	router := newRouter(c, suite.app)
	suite.testServer = httptest.NewServer(router)

	if _, err := suite.app.AdminService.Create(models.Test_admin); err != nil {
		log.Fatal().Msgf("error creating test admin: %v", err.Error())
	}

	if _, err := suite.app.ResidentService.Create(models.Test_resident); err != nil {
		log.Fatal().Msgf("error creating test resident: %v", err.Error())
	}
}

func (suite *authRouterSuite) TearDownSuite() {
	defer suite.testServer.Close()

	if err := suite.container.Terminate(context.Background()); err != nil {
		require.NoError(suite.T(), fmt.Errorf("error tearing down container: %v", err))
	}
}

func (suite *authRouterSuite) TestLogin_Admin_Positive() {
	requestBody := []byte(fmt.Sprintf(`{
    "id":"%s",
    "password":"%s"
  }`,
		models.Test_admin.ID,
		models.Test_admin.Password))
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

	expectedUser := models.Test_admin.AsUser()
	suite.Empty(cmp.Diff(expectedUser, session.User), "user in response did not equal expected user")

	err = checkAccessToken(suite.app.JWTService, session.AccessToken, expectedUser)
	suite.NoError(err)

	err = checkRefreshToken(suite.app.JWTService, response.Cookies(), expectedUser.ID, expectedUser.TokenVersion)
	suite.NoError(err)
}

func (suite *authRouterSuite) TestLogin_Resident_Positive() {
	requestBody := []byte(fmt.Sprintf(`{
    "id":"%s",
    "password":"%s"
  }`, models.Test_resident.ID, models.Test_resident.Password))
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

	expectedUser := models.Test_resident.AsUser()
	suite.Empty(cmp.Diff(expectedUser, session.User), "user in response did not equal expected user")

	err = checkAccessToken(suite.app.JWTService, session.AccessToken, expectedUser)
	suite.NoError(err)

	err = checkRefreshToken(suite.app.JWTService, response.Cookies(), expectedUser.ID, expectedUser.TokenVersion)
	suite.NoError(err)
}

func (suite *authRouterSuite) TestRefreshTokens_Positive() {
	request, err := http.NewRequest("POST", suite.testServer.URL+"/api/refresh-tokens", bytes.NewBuffer([]byte{}))
	if err != nil {
		suite.NoError(fmt.Errorf("error creating http.Request: %s", err))
		return
	}

	expectedUser := models.Test_resident.AsUser()
	refreshToken, err := suite.app.JWTService.NewRefresh(expectedUser)
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
