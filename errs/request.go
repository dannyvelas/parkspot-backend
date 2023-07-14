package errs

import (
	"fmt"
	"net/http"
)

var (
	Unauthorized   = NewApiErr(http.StatusUnauthorized, "unauthorized")
	NotFound       = NewApiErr(http.StatusNotFound, "not found")
	MissingIDField = NewApiErr(http.StatusBadRequest, "ID field is required but missing")
	IDNotUUID      = NewApiErr(http.StatusBadRequest, "ID field is not a UUID")
	InvalidResID   = NewApiErr(http.StatusBadRequest, "ResidentID must start be a 'B' or a 'T', followed by 7 numbers")
	AlreadyExists  = NewApiErr(http.StatusBadRequest, "already exists")
)

func BadRequest(message string) *ApiErr {
	return NewApiErr(http.StatusBadRequest, message)
}

func NewNotFound(resource string) *ApiErr {
	return &ApiErr{http.StatusNotFound, fmt.Errorf("%s %w", resource, NotFound)}
}

func NewUnauthorized(message string) *ApiErr {
	return &ApiErr{http.StatusUnauthorized, fmt.Errorf("%s %w", message, Unauthorized)}
}

func EmptyFields(fields string) *ApiErr {
	return NewApiErr(http.StatusBadRequest, "One or more missing fields: "+fields)
}

func InvalidFields(fields string) *ApiErr {
	return NewApiErr(http.StatusBadRequest, "One or more invalid fields: "+fields)
}

func Malformed(payload string) *ApiErr {
	return NewApiErr(http.StatusBadRequest, payload+" malformed")
}

func NewAlreadyExists(resource string) *ApiErr {
	return &ApiErr{http.StatusBadRequest, fmt.Errorf("%s %w", resource, AlreadyExists)}
}

func AllEditFieldsEmpty(fields string) *ApiErr {
	return NewApiErr(http.StatusBadRequest, fmt.Sprintf("All edit fields (%s) cannot be empty", fields))
}
