package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// httpRequestDuration computes latency and request volume
var httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "parkspot_http_request_duration_seconds",
	Help:    "Latency of HTTP requests, by route, method, and status class.",
	Buckets: prometheus.DefBuckets,
}, []string{"route", "method", "status"})

// metricsMiddleware records httpRequestDuration for every request except
// /up: kamal-proxy hits it every second (proxy.healthcheck.interval: 1 in
// deploy.yml), which would swamp every real route's numbers.
func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/up" {
			next.ServeHTTP(w, r)
			return
		}

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()
		next.ServeHTTP(ww, r)

		httpRequestDuration.WithLabelValues(metricsRoutePattern(r), r.Method, statusClass(ww.Status())).
			Observe(time.Since(start).Seconds())
	})
}

// metricsRoutePattern returns chi's matched route pattern, or "unmatched"
// for a 404 - never the raw request path. Falling back to the raw path
// (like respond.go's routePattern does, safely, since it only ever runs
// for errors on routes that already matched) would give an unbounded
// label here, since every request, including junk paths bots probe, goes
// through this middleware.
func metricsRoutePattern(r *http.Request) string {
	rctx := chi.RouteContext(r.Context())
	if rctx == nil {
		return "unmatched"
	}

	pattern := rctx.RoutePattern()
	if pattern == "" {
		return "unmatched"
	}

	return pattern
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
