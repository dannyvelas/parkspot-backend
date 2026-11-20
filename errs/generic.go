package errs

import (
	"fmt"
	"net/http"
)

var (
	Unauthorized   = NewAPIErr(http.StatusUnauthorized, "unauthorized")
	MissingIDField = NewAPIErr(http.StatusBadRequest, "ID field is required but missing")
	IDNotUUID      = NewAPIErr(http.StatusBadRequest, "ID field is not a UUID")
	InvalidResID   = NewAPIErr(http.StatusBadRequest, "ResidentID must start be a 'B' or a 'T', followed by 7 numbers")
)

func NewUnauthorized(message string) *APIErr {
	return &APIErr{http.StatusUnauthorized, fmt.Errorf("%s %w", message, Unauthorized)}
}

func EmptyFields(fields string) *APIErr {
	return NewAPIErr(http.StatusBadRequest, "One or more missing fields: "+fields)
}

func InvalidFields(fields string) *APIErr {
	return NewAPIErr(http.StatusBadRequest, "One or more invalid fields: "+fields)
}

func AllEditFieldsEmpty(fields string) *APIErr {
	return NewAPIErr(http.StatusBadRequest, fmt.Sprintf("All edit fields (%s) cannot be empty", fields))
}
