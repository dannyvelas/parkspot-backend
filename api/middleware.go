package api

import (
	"github.com/dannyvelas/lasvistas_api/app"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
)

type middleware struct {
	jwtService app.JWTService
}

func newMiddleware(jwtService app.JWTService) middleware {
	return middleware{
		jwtService: jwtService,
	}
}

func (m middleware) authenticate(firstRole models.Role, roles ...models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				log.Debug().Msg("No 'Authorization' header was present with 'Bearer ' prefix.")
				respondError(w, errUnauthorized)
				return
			}

			accessToken := strings.TrimPrefix(authHeader, "Bearer ")
			accessPayload, err := m.jwtService.ParseAccess(accessToken)
			if err != nil {
				log.Debug().Msgf("Error parsing: %v", err)
				respondError(w, errUnauthorized)
				return
			}

			permittedRoles := append([]models.Role{firstRole}, roles...)
			userHasPermittedRole := func() bool {
				for _, role := range permittedRoles {
					if accessPayload.Role == role {
						return true
					}
				}
				return false
			}()
			if !userHasPermittedRole {
				log.Debug().Msgf("User role: %s, not in permittedRoles: %v", accessPayload.Role, permittedRoles)
				respondError(w, errUnauthorized)
				return
			}

			ctx := r.Context()
			updatedCtx := ctxWithAccessPayload(ctx, accessPayload)
			updatedReq := r.WithContext(updatedCtx)

			next.ServeHTTP(w, updatedReq)
		})
	}
}
