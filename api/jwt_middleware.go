package api

import (
	"errors"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"time"
)

var (
	errNotSigningMethodHMAC = errors.New("jwt: Not using SigningMethodHMAC")
	errCastingJWTClaims     = errors.New("jwt: Failed to cast JWT token to JWTClaims struct")
	errInvalidToken         = errors.New("jwt: Invalid Token")
)

type jwtClaims struct {
	User user `json:"user"`
	jwt.StandardClaims
}

type jwtMiddleware struct {
	tokenSecret []byte
}

func NewJWTMiddleware(tokenConfig config.TokenConfig) jwtMiddleware {
	return jwtMiddleware{tokenSecret: []byte(tokenConfig.Secret())}
}

func (jwtMiddleware jwtMiddleware) newJWT(id string, firstName string, lastName string, email string, role Role) (string, error) {
	claims := jwtClaims{
		user{id, firstName, lastName, email, role},
		jwt.StandardClaims{ExpiresAt: time.Now().Add(time.Minute * 15).Unix()},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtMiddleware.tokenSecret)
}

func (jwtMiddleware jwtMiddleware) parseJWT(tokenString string) (user, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errNotSigningMethodHMAC
		}

		return jwtMiddleware.tokenSecret, nil
	})
	if err != nil {
		return user{}, err
	}

	if claims, ok := token.Claims.(*jwtClaims); !ok {
		return user{}, errCastingJWTClaims
	} else if !token.Valid {
		return user{}, errInvalidToken
	} else {
		return claims.User, nil
	}
}

func (jwtMiddleware jwtMiddleware) Authenticate(role Role, roles ...Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("jwt")
			if err != nil {
				respondError(w, errUnauthorized)
				return
			}

			user, err := jwtMiddleware.parseJWT(cookie.Value)
			if err != nil {
				respondError(w, errUnauthorized)
				return
			}

			permittedRoles := append([]Role{role}, roles...)
			userHasPermittedRole := func() bool {
				for _, role := range permittedRoles {
					if user.Role == role {
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
			updatedCtx := ctxWithUser(ctx, user)
			updatedReq := r.WithContext(updatedCtx)

			next.ServeHTTP(w, updatedReq)
		})
	}
}
