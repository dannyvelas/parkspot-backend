package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// metricsMiddleware records httpRequestDuration
func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/up" {
			next.ServeHTTP(w, r)
			return
		}

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()
		next.ServeHTTP(ww, r)

		// default to unmatched bc every request, including junk, goes through here. otherwise, we explode cardinality
		route := routePatternOr(r, "unmatched")
		httpRequestDuration.WithLabelValues(route, metricsMethod(r.Method), statusClass(ww.Status())).
			Observe(time.Since(start).Seconds())
	})
}

// metricsMethod just makes sure method is one of the methods this backend uses. protects cardinality
func metricsMethod(method string) string {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions:
		return method
	default:
		return "other"
	}
}

func statusClass(status int) string {
	switch {
	case status >= 500:
		return "5xx"
	case status >= 400:
		return "4xx"
	case status >= 300:
		return "3xx"
	case status >= 200:
		return "2xx"
	case status >= 100:
		return "1xx"
	default:
		return "unknown"
	}
}
