package app

import (
	"errors"
	"time"

	"github.com/dannyvelas/parkspot-backend/config"
	"github.com/dannyvelas/parkspot-backend/errs"
	"github.com/dannyvelas/parkspot-backend/models"
	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrNotSigningMethodHMAC = errors.New("jwt: Not using SigningMethodHMAC")
	ErrCastingJWTClaims     = errs.NewSentinel("ErrCastingJWTClaims", "jwt: Failed to cast JWT to JWTClaims struct")
	ErrInvalidToken         = errs.NewSentinel("ErrInvalidToken", "jwt: Invalid Token")
)

type JWTService struct {
	accessSecret  []byte
	refreshSecret []byte
}

func NewJWTService(tokenConfig config.TokenConfig) JWTService {
	return JWTService{
		accessSecret:  []byte(tokenConfig.AccessSecret),
		refreshSecret: []byte(tokenConfig.RefreshSecret),
	}
}

type accessClaims struct {
	Payload AccessPayload `json:"payload"`
	jwt.RegisteredClaims
}

type AccessPayload struct {
	ID   string      `json:"id"`
	Role models.Role `json:"role"`
}

type refreshClaims struct {
	User models.User `json:"user"`
	jwt.RegisteredClaims
}

func (jwtService JWTService) NewAccess(id string, role models.Role) (string, error) {
	claims := accessClaims{
		AccessPayload{id, role},
		jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15))},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtService.accessSecret)
}

func (jwtService JWTService) NewRefresh(user models.User) (string, error) {
	claims := refreshClaims{
		user,
		jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().AddDate(0, 0, 7))}, // 7 days
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtService.refreshSecret)
}

func (jwtService JWTService) ParseAccess(tokenString string) (AccessPayload, error) {
	token, err := jwt.ParseWithClaims(tokenString, &accessClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrNotSigningMethodHMAC
		}

		return jwtService.accessSecret, nil
	})
	if err != nil {
		return AccessPayload{}, errs.Unauthorized
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
	token, err := jwt.ParseWithClaims(tokenString, &refreshClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrNotSigningMethodHMAC
		}

		return jwtService.refreshSecret, nil
	})
	if err != nil {
		return models.User{}, errs.Unauthorized
	}

	if claims, ok := token.Claims.(*refreshClaims); !ok {
		return models.User{}, ErrCastingJWTClaims
	} else if !token.Valid {
		return models.User{}, ErrInvalidToken
	} else {
		return claims.User, nil
	}
}
