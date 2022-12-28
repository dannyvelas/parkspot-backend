package api

import (
	"github.com/dannyvelas/lasvistas_api/app"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/util"
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
				respondError(w, errs.Unauthorized)
				return
			}

			accessToken := strings.TrimPrefix(authHeader, "Bearer ")
			accessPayload, err := m.jwtService.ParseAccess(accessToken)
			if err != nil {
				respondError(w, errs.Unauthorized)
				return
			}

			permittedRoles := append([]models.Role{firstRole}, roles...)
			userHasPermittedRole := util.Contains(permittedRoles, accessPayload.Role)
			if !userHasPermittedRole {
				log.Debug().Msgf("User role: %s, not in permittedRoles: %v", accessPayload.Role, permittedRoles)
				respondError(w, errs.Unauthorized)
				return
			}

			ctx := r.Context()
			updatedCtx := ctxWithAccessPayload(ctx, accessPayload)
			updatedReq := r.WithContext(updatedCtx)

			next.ServeHTTP(w, updatedReq)
		})
	}
}
