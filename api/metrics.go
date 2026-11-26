package api

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var unhandledErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "parkspot_unhandled_errors_total",
	Help: "Count of errors that reached respondError's unhandled branch and were returned to the client as a 500.",
}, []string{"route", "error"})

var httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "parkspot_http_request_duration_seconds",
	Help:    "Latency of HTTP requests, by route, method, and status class.",
	Buckets: prometheus.DefBuckets,
}, []string{"route", "method", "status"})
