package api

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var unhandledErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
	Name: "parkspot_unhandled_errors_total",
	Help: "Count of errors that reached respondError's unhandled branch and were returned to the client as a 500.",
})
