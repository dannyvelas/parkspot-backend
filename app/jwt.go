package app

import (
	"errors"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

var (
	ErrNotSigningMethodHMAC = errors.New("jwt: Not using SigningMethodHMAC")
	ErrCastingJWTClaims     = errors.New("jwt: Failed to cast JWT to JWTClaims struct")
	ErrInvalidToken         = errors.New("jwt: Invalid Token")
)

type JWTService struct {
	accessSecret  []byte
	refreshSecret []byte
}

func NewJWTService(tokenConfig config.TokenConfig) JWTService {
	return JWTService{
		accessSecret:  []byte(tokenConfig.AccessSecret()),
		refreshSecret: []byte(tokenConfig.RefreshSecret()),
	}
}

type accessClaims struct {
	Payload AccessPayload `json:"payload"`
	jwt.StandardClaims
}

type AccessPayload struct {
	ID   string      `json:"id"`
	Role models.Role `json:"role"`
}

type refreshClaims struct {
	User models.User `json:"user"`
	jwt.StandardClaims
}

func (jwtService JWTService) NewAccess(id string, role models.Role) (string, error) {
	claims := accessClaims{
		AccessPayload{id, role},
		jwt.StandardClaims{ExpiresAt: time.Now().Add(time.Minute * 15).Unix()},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtService.accessSecret)
}

func (jwtService JWTService) NewRefresh(user models.User) (string, error) {
	claims := refreshClaims{
		user,
		jwt.StandardClaims{ExpiresAt: time.Now().AddDate(0, 0, 7).Unix()}, // 7 days
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtService.refreshSecret)
}

func (jwtService JWTService) ParseAccess(tokenString string) (AccessPayload, error) {
	token, err := jwt.ParseWithClaims(tokenString, &accessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrNotSigningMethodHMAC
		}

		return jwtService.accessSecret, nil
	})
	if err != nil {
		return AccessPayload{}, err
	}

	if claims, ok := token.Claims.(*accessClaims); !ok {
		return AccessPayload{}, ErrCastingJWTClaims
	} else if !token.Valid {
		return AccessPayload{}, ErrInvalidToken
	} else {
		return claims.Payload, nil
	}
}

func (jwtService JWTService) ParseRefresh(tokenString string) (models.User, error) {
	token, err := jwt.ParseWithClaims(tokenString, &refreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrNotSigningMethodHMAC
		}

		return jwtService.refreshSecret, nil
	})
	if err != nil {
		return models.User{}, err
	}

	if claims, ok := token.Claims.(*refreshClaims); !ok {
		return models.User{}, ErrCastingJWTClaims
	} else if !token.Valid {
		return models.User{}, ErrInvalidToken
	} else {
		return claims.User, nil
	}
}
