package routing

import (
	"context"
	"github.com/dannyvelas/parkspot-api/routing/internal"
	"github.com/rs/zerolog/log"
	"net/http"
)

func (authenticator RoutingAuthenticator) Authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("jwt")
		if err != nil {
			log.Debug().Msg("Cookie not found!")
			internal.HandleError(w, internal.Unauthorized)
			return
		}

		userId, err := authenticator.ParseJWT(cookie.Value)
		if err != nil {
			log.Debug().Msg("Error parsing payload: " + err.Error())
			internal.HandleError(w, internal.Unauthorized)
			return
		}

		ctx := r.Context()
		updatedCtx := context.WithValue(ctx, "id", userId)
		updatedReq := r.WithContext(updatedCtx)

		next.ServeHTTP(w, updatedReq)
	})
}
