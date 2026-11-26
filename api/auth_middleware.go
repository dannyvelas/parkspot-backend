package api

import (
	"net/http"
	"slices"
	"strings"

	"github.com/dannyvelas/parkspot-backend/app"
	"github.com/dannyvelas/parkspot-backend/errs"
	"github.com/dannyvelas/parkspot-backend/models"
	"github.com/rs/zerolog/log"
)

type authMiddleware struct {
	jwtService app.JWTService
}

func newAuthMiddleware(jwtService app.JWTService) authMiddleware {
	return authMiddleware{
		jwtService: jwtService,
	}
}

func (m authMiddleware) authenticate(firstRole models.Role, roles ...models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				respondError(w, r, errs.Unauthorized)
				return
			}

			accessToken := strings.TrimPrefix(authHeader, "Bearer ")
			accessPayload, err := m.jwtService.ParseAccess(accessToken)
			if err != nil {
				respondError(w, r, err)
				return
			}

			permittedRoles := append([]models.Role{firstRole}, roles...)
			userHasPermittedRole := slices.Contains(permittedRoles, accessPayload.Role)
			if !userHasPermittedRole {
				log.Debug().Msgf("User role: %s, not in permittedRoles: %v", accessPayload.Role, permittedRoles)
				respondError(w, r, errs.Unauthorized)
				return
			}

			ctx := r.Context()
			updatedCtx := ctxWithAccessPayload(ctx, accessPayload)
			updatedReq := r.WithContext(updatedCtx)

			next.ServeHTTP(w, updatedReq)
		})
	}
}
