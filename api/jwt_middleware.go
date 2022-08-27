package api

import (
	"errors"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"strings"
	"time"
)

var (
	errNotSigningMethodHMAC = errors.New("jwt: Not using SigningMethodHMAC")
	errCastingJWTClaims     = errors.New("jwt: Failed to cast JWT to JWTClaims struct")
	errInvalidToken         = errors.New("jwt: Invalid Token")
)

type jwtMiddleware struct {
	accessSecret  []byte
	refreshSecret []byte
}

func NewJWTMiddleware(tokenConfig config.TokenConfig) jwtMiddleware {
	return jwtMiddleware{
		accessSecret:  []byte(tokenConfig.AccessSecret()),
		refreshSecret: []byte(tokenConfig.RefreshSecret()),
	}
}

type accessClaims struct {
	Payload accessPayload `json:"payload"`
	jwt.StandardClaims
}

type accessPayload struct {
	Id   string `json:"id"`
	Role role   `json:"role"`
}

func (jwtMiddleware jwtMiddleware) newAccess(id string, role role) (string, error) {
	claims := accessClaims{
		accessPayload{id, role},
		jwt.StandardClaims{ExpiresAt: time.Now().Add(time.Minute * 15).Unix()},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtMiddleware.accessSecret)
}

type refreshClaims struct {
	Payload refreshPayload `json:"payload"`
	jwt.StandardClaims
}

type refreshPayload struct {
	Id      string `json:"id"`
	Version int    `json:"version"`
}

func (jwtMiddleware jwtMiddleware) newRefresh(id string, version int) (string, error) {
	claims := refreshClaims{
		refreshPayload{id, version},
		jwt.StandardClaims{ExpiresAt: time.Now().AddDate(1, 0, 0).Unix()}, // 1y from now
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtMiddleware.accessSecret)
}

func (jwtMiddleware jwtMiddleware) parseAccess(tokenString string) (accessPayload, error) {
	token, err := jwt.ParseWithClaims(tokenString, &accessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errNotSigningMethodHMAC
		}

		return jwtMiddleware.accessSecret, nil
	})
	if err != nil {
		return accessPayload{}, err
	}

	if claims, ok := token.Claims.(*accessClaims); !ok {
		return accessPayload{}, errCastingJWTClaims
	} else if !token.Valid {
		return accessPayload{}, errInvalidToken
	} else {
		return claims.Payload, nil
	}
}

func (jwtMiddleware jwtMiddleware) parseRefresh(tokenString string) (refreshPayload, error) {
	token, err := jwt.ParseWithClaims(tokenString, &refreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errNotSigningMethodHMAC
		}

		return jwtMiddleware.refreshSecret, nil
	})
	if err != nil {
		return refreshPayload{}, err
	}

	if claims, ok := token.Claims.(*refreshClaims); !ok {
		return refreshPayload{}, errCastingJWTClaims
	} else if !token.Valid {
		return refreshPayload{}, errInvalidToken
	} else {
		return claims.Payload, nil
	}
}

func (jwtMiddleware jwtMiddleware) authenticate(firstRole role, roles ...role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				respondError(w, errUnauthorized)
				return
			}

			accessToken := strings.TrimPrefix(authHeader, "Bearer ")
			accessPayload, err := jwtMiddleware.parseAccess(accessToken)
			if err != nil {
				respondError(w, errUnauthorized)
				return
			}

			permittedRoles := append([]role{firstRole}, roles...)
			userHasPermittedRole := func() bool {
				for _, role := range permittedRoles {
					if accessPayload.Role == role {
						return true
					}
				}
				return false
			}()
			if !userHasPermittedRole {
				respondError(w, errUnauthorized)
				return
			}

			ctx := r.Context()
			updatedCtx := ctxWithAccessPayload(ctx, accessPayload)
			updatedReq := r.WithContext(updatedCtx)

			next.ServeHTTP(w, updatedReq)
		})
	}
}
