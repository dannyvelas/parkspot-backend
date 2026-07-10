package errs

import (
	"fmt"
	"net/http"
)

var (
	Unauthorized   = NewAPIErr(http.StatusUnauthorized, "unauthorized")
	NotFound       = NewAPIErr(http.StatusNotFound, "not found")
	MissingIDField = NewAPIErr(http.StatusBadRequest, "ID field is required but missing")
	IDNotUUID      = NewAPIErr(http.StatusBadRequest, "ID field is not a UUID")
	InvalidResID   = NewAPIErr(http.StatusBadRequest, "ResidentID must start be a 'B' or a 'T', followed by 7 numbers")
	AlreadyExists  = NewAPIErr(http.StatusBadRequest, "already exists")
)

func BadRequest(message string) *APIErr {
	return NewAPIErr(http.StatusBadRequest, message)
}

func NewNotFound(resource string) *APIErr {
	return &APIErr{http.StatusNotFound, fmt.Errorf("%s %w", resource, NotFound)}
}

func NewUnauthorized(message string) *APIErr {
	return &APIErr{http.StatusUnauthorized, fmt.Errorf("%s %w", message, Unauthorized)}
}

func EmptyFields(fields string) *APIErr {
	return NewAPIErr(http.StatusBadRequest, "One or more missing fields: "+fields)
}

func InvalidFields(fields string) *APIErr {
	return NewAPIErr(http.StatusBadRequest, "One or more invalid fields: "+fields)
}

func Malformed(payload string) *APIErr {
	return NewAPIErr(http.StatusBadRequest, payload+" malformed")
}

func NewAlreadyExists(resource string) *APIErr {
	return &APIErr{http.StatusBadRequest, fmt.Errorf("%s %w", resource, AlreadyExists)}
}

func AllEditFieldsEmpty(fields string) *APIErr {
	return NewAPIErr(http.StatusBadRequest, fmt.Sprintf("All edit fields (%s) cannot be empty", fields))
}
