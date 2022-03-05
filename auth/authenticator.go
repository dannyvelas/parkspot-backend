package auth

import (
	"github.com/dannyvelas/parkspot-api/config"
)

type Authenticator struct {
	tokenSecret []byte
}

func NewAuthenticator(tokenConfig config.TokenConfig) Authenticator {
	return Authenticator{tokenSecret: []byte(tokenConfig.Secret())}
}
