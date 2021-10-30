package auth

import (
	"github.com/dannyvelas/parkspot-api/config"
	"net/http"
)

type Authenticator struct {
	tokenSecret []byte
}

func NewAuthenticator(tokenConfig config.TokenConfig) *Authenticator {
	return &Authenticator{tokenSecret: []byte(tokenConfig.Secret())}
}

func (authenticator Authenticator) AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//cookie, err := r.Cookie("jwt")
		//if err != nil {
		//	http.Error(w, "Unauthorized", http.StatusUnauthorized)
		//	return
		//}

		//jwtPayload, err := authenticator.parseJWT(cookie.Value)
		//if err != nil {
		//	http.Error(w, "Unauthorized", http.StatusUnauthorized)
		//	return
		//}
		//next.ServeHTTP(w, r)
	})
}
