package api

import (
	"errors"

	"github.com/dannyvelas/parkspot-backend/errs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var unhandledErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "parkspot_unhandled_errors_total",
	Help: "Count of errors that reached respondError's unhandled branch and were returned to the client as a 500.",
}, []string{"route", "error"})

func errorLabel(err error) string {
	var sentinel *errs.Sentinel
	if errors.As(err, &sentinel) {
		return sentinel.Label()
	}
	return "unknown"
}
