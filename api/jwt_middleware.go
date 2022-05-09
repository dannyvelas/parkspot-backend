package api

import (
	"context"
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

type jwtPayload struct {
	Id   string `json:"id"`
	Role Role   `json:"role"`
}

type jwtClaims struct {
	jwtPayload
	jwt.StandardClaims
}

type JWTMiddleware struct {
	tokenSecret []byte
}

func NewJWTMiddleware(tokenConfig config.TokenConfig) JWTMiddleware {
	return JWTMiddleware{tokenSecret: []byte(tokenConfig.Secret())}
}

func (jwtMiddleware JWTMiddleware) newJWT(id string, role Role) (string, error) {
	claims := jwtClaims{
		jwtPayload{
			id,
			role,
		},
		jwt.StandardClaims{ExpiresAt: time.Now().Add(time.Minute * 15).Unix()},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtMiddleware.tokenSecret)
}

func (jwtMiddleware JWTMiddleware) parseJWT(tokenString string) (jwtPayload, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errNotSigningMethodHMAC
		}

		return jwtMiddleware.tokenSecret, nil
	})
	if err != nil {
		return jwtPayload{}, err
	}

	if claims, ok := token.Claims.(*jwtClaims); !ok {
		return jwtPayload{}, errCastingJWTClaims
	} else if !token.Valid {
		return jwtPayload{}, errInvalidToken
	} else {
		return claims.jwtPayload, nil
	}
}

func (jwtMiddleware JWTMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("jwt")
		if err != nil {
			respondError(w, errUnauthorized)
			return
		}

		userId, err := jwtMiddleware.parseJWT(cookie.Value)
		if err != nil {
			respondError(w, errUnauthorized)
			return
		}

		ctx := r.Context()
		updatedCtx := context.WithValue(ctx, "id", userId)
		updatedReq := r.WithContext(updatedCtx)

		next.ServeHTTP(w, updatedReq)
	})
}
