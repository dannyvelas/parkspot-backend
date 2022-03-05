package auth

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"time"
)

type JWTClaims struct {
	Id string `json:"id"`
	jwt.StandardClaims
}

func (authenticator Authenticator) NewJWT(id string) (string, error) {
	claims := JWTClaims{
		id,
		jwt.StandardClaims{ExpiresAt: time.Now().Add(time.Minute * 15).Unix()},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(authenticator.tokenSecret)
}

func (authenticator Authenticator) ParseJWT(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Not using SigningMethodHMAC!")
		}

		return authenticator.tokenSecret, nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*JWTClaims); !ok || !token.Valid {
		if !ok {
			return "", errors.New("Failure casting JWTClaims!")
		} else {
			return "", errors.New("Token not valid!")
		}
	} else {
		return claims.Id, nil
	}
}
