package api

import (
	"net/http"
)

type apiError interface {
	apiError() (int, string)
}

type sentinelError struct {
	statusCode int
	message    string
}

var (
	errUnauthorized        = sentinelError{http.StatusUnauthorized, "Unauthorized"}
	errBadRequest          = sentinelError{http.StatusBadRequest, "Bad Request"}
	errEmptyFields         = sentinelError{http.StatusBadRequest, "One or more missing fields"}
	errInvalidFields       = sentinelError{http.StatusBadRequest, "One or more invalid fields"}
	errNotFound            = sentinelError{http.StatusNotFound, "Not Found"}
	errInternalServerError = sentinelError{http.StatusInternalServerError, "Internal Server Error"}
)

func (e sentinelError) Error() string {
	return e.message
}

func (e sentinelError) apiError() (int, string) {
	return e.statusCode, e.message
}
