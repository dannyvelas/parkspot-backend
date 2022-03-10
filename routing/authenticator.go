package routing

import (
	"github.com/dannyvelas/lasvistas_api/auth"
)

type RoutingAuthenticator struct {
	auth.Authenticator
}

func NewAuthenticator(authenticator auth.Authenticator) RoutingAuthenticator {
	return RoutingAuthenticator{
		Authenticator: authenticator,
	}
}
