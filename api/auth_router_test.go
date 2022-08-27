package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/storage"
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
	testServer    *httptest.Server
	jwtMiddleware jwtMiddleware
	adminJWT      string
}

func TestAuthRouter(t *testing.T) {
	suite.Run(t, new(authRouterSuite))
}

func (suite *authRouterSuite) SetupSuite() {
	c, err := config.NewConfig()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	database, err := storage.NewDatabase(c.Postgres())
	if err != nil {
		log.Fatal().Msgf("Failed to start database: %v", err)
	}

	suite.testServer = newTestServer(c, storage.NewRepos(database))

	suite.jwtMiddleware = NewJWTMiddleware(c.Token())

	suite.adminJWT, err = suite.jwtMiddleware.newAccess("some-uuid", AdminRole)
	if err != nil {
		log.Fatal().Msgf("Failed to create JWT: %v", err)
	}

	err = createTestResidents(suite.testServer.URL, suite.adminJWT)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
}

func (suite authRouterSuite) TearDownSuite() {
	defer suite.testServer.Close()

	err := deleteTestResidents(suite.testServer.URL, suite.adminJWT)
	if err != nil {
		log.Error().Msg("auth_router_test.TearDownSuite: " + err.Error())
		return
	}
}

func (suite authRouterSuite) TestLogin_Admin_Positive() {
	requestBody := []byte(`{
    "id":"test",
    "password":"notapassword"
  }`)
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

	var loginResponse loginResponse
	if err := json.NewDecoder(response.Body).Decode(&loginResponse); err != nil {
		suite.NoError(err)
		return
	}

	// check user
	expectedUser := newUser("test",
		"Daniel",
		"Velasquez",
		"email@example.com",
		AdminRole,
		0)
	suite.Empty(cmp.Diff(expectedUser, loginResponse.User), "response body was not the same")

	err = checkAccessToken(suite.jwtMiddleware, loginResponse.AccessToken, expectedUser.Id, expectedUser.Role)
	suite.NoError(err)

	err = checkRefreshToken(suite.jwtMiddleware, response.Cookies(), expectedUser.Id, expectedUser.TokenVersion)
	suite.NoError(err)
}

func (suite authRouterSuite) TestLogin_Resident_Positive() {
	requestBody := []byte(fmt.Sprintf(`{
    "id":"%s",
    "password":"%s"
  }`, testResident.Id, testResident.Password))
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

	var loginResponse loginResponse
	if err := json.NewDecoder(response.Body).Decode(&loginResponse); err != nil {
		suite.NoError(err)
		return
	}

	// check user
	expectedUser := newUser(testResident.Id,
		testResident.FirstName,
		testResident.LastName,
		testResident.Email,
		ResidentRole,
		0)
	suite.Empty(cmp.Diff(expectedUser, loginResponse.User), "response body was not the same")

	err = checkAccessToken(suite.jwtMiddleware, loginResponse.AccessToken, expectedUser.Id, expectedUser.Role)
	suite.NoError(err)

	err = checkRefreshToken(suite.jwtMiddleware, response.Cookies(), expectedUser.Id, expectedUser.TokenVersion)
	suite.NoError(err)
}

func (suite authRouterSuite) TestCreate_ResidentDuplicateEmail_Negative() {
	requestBody := []byte(fmt.Sprintf(`{
    "residentId": "B0000000",
    "firstName": "first",
    "lastName": "last",
    "phone": "123456789",
    "email": "%s",
    "password":"password",
    "unlimDays": false
  }`,
		testResident.Email))

	responseBody, statusCode, err := authenticatedReq("POST", suite.testServer.URL+"/api/account", requestBody, suite.adminJWT)
	if err != nil {
		suite.NoError(fmt.Errorf("error sending request: %v", err))
		return
	}
	defer responseBody.Close()

	if statusCode == http.StatusOK {
		suite.NoError(fmt.Errorf("Successfully created resident with duplicate email when it shouldn't have"))
		return
	}

	suite.Equal(http.StatusBadRequest, statusCode, "Expected bad request got: %d", statusCode)

	bodyBytes, err := io.ReadAll(responseBody)
	if err != nil {
		suite.NoError(fmt.Errorf("error getting error response: %v", err))
		return
	}

	suite.Contains(string(bodyBytes), "email") // assert bad request happened bc of email
}

// helpers
// accessToken, id, role
// cookies, jwtMiddleware,
func checkAccessToken(jwtMiddleware jwtMiddleware, token string, expectedID string, expectedRole role) error {
	if token == "" {
		return fmt.Errorf("accessToken was empty")
	} else if payload, err := jwtMiddleware.parseAccess(token); err != nil {
		return fmt.Errorf("Error parsing access token (%s): %v", token, err)
	} else if expectedID != payload.Id {
		return fmt.Errorf("user id (%s) was not the same to access payload id (%s)", expectedID, payload.Id)
	} else if expectedRole != payload.Role {
		return fmt.Errorf("user role (%v) was not the same to access payload role (%v)", expectedRole, payload.Role)
	}

	return nil
}

func checkRefreshToken(jwtMiddleware jwtMiddleware, cookies []*http.Cookie, expectedID string, expectedVersion int) error {
	refreshCookie := func() *http.Cookie {
		for _, cookie := range cookies {
			if cookie.Name == refreshCookieKey {
				return cookie
			}
		}
		return nil
	}()
	if refreshCookie == nil {
		return fmt.Errorf("cookie with key of %s not found", refreshCookieKey)
	}

	if payload, err := jwtMiddleware.parseRefresh(refreshCookie.Value); err != nil {
		return fmt.Errorf("Error parsing refresh token (%s): %v", refreshCookie.Value, err)
	} else if expectedID != payload.Id {
		return fmt.Errorf("user id (%s) was not the same to refresh payload id (%s)", expectedID, payload.Id)
	} else if expectedVersion != payload.Version {
		return fmt.Errorf("user version (%v) was not the same to refresh payload version (%v)", expectedVersion, payload.Version)
	}

	return nil
}
