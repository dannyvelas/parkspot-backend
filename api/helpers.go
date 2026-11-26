package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// routePatternOr returns chi's matched route pattern, or fallback if nothing
// matched (e.g. a 404). Shared by routePattern above and metricsMiddleware,
// which need different fallbacks - see metricsMiddleware's comment.
func routePatternOr(r *http.Request, fallback string) string {
	rctx := chi.RouteContext(r.Context())
	if rctx == nil {
		return fallback
	}

	pattern := rctx.RoutePattern()
	if pattern == "" {
		return fallback
	}

	return pattern
}
