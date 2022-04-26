package api

import (
	"net/http"
)

type responseError struct {
	statusCode int
	message    string
}

var (
	errUnauthorized        = responseError{http.StatusUnauthorized, "Unauthorized"}
	errBadRequest          = responseError{http.StatusBadRequest, "Bad Request"}
	errEmptyFields         = responseError{http.StatusBadRequest, "One or more missing fields"}
	errInvalidFields       = responseError{http.StatusBadRequest, "One or more invalid fields"}
	errNotFound            = responseError{http.StatusNotFound, "Not Found"}
	errInternalServerError = responseError{http.StatusInternalServerError, "Internal Server Error"}
)

func (e responseError) Error() string {
	return e.message
}
