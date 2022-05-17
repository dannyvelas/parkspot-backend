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
	errEmptyFields         = responseError{http.StatusBadRequest, "One or more missing fields"}
	errInvalidFields       = responseError{http.StatusBadRequest, "One or more invalid fields"}
	errInternalServerError = responseError{http.StatusInternalServerError, "Internal Server Error"}
)

func (e responseError) Error() string {
	return e.message
}

func newErrNotFound(resource string) responseError {
	return responseError{http.StatusNotFound, resource + " not found"}
}

func newErrBadRequest(message string) responseError {
	return responseError{http.StatusBadRequest, message}
}

func newErrMalformed(payload string) responseError {
	return responseError{http.StatusBadRequest, payload + " malformed"}
}
