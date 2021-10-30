package auth

import (
	"errors"
	"github.com/golang-jwt/jwt"
)

type JWTPayload struct {
	Id string `json:"id"`
}

type jwtClaims struct {
	JWTPayload
	jwt.StandardClaims
}

func (authenticator Authenticator) NewJWT(id string) (string, error) {
	claims := jwtClaims{
		JWTPayload{Id: id},
		jwt.StandardClaims{ExpiresAt: 15000},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(authenticator.tokenSecret)
}

func (authenticator Authenticator) parseJWT(tokenString string) (JWTPayload, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Unauthorized")
		}

		return authenticator.tokenSecret, nil
	})
	if err != nil {
		return JWTPayload{}, err
	}

	if claims, ok := token.Claims.(jwtClaims); ok && token.Valid {
		return claims.JWTPayload, nil
	} else {
		return JWTPayload{}, errors.New("Unauthorized")
	}
}
