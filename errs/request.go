package errs

import (
	"fmt"
	"net/http"
)

var (
	Unauthorized = NewApiErr(http.StatusUnauthorized, "unauthorized")
	NotFound     = NewApiErr(http.StatusNotFound, "not found")
)

func BadRequest(message string) *ApiErr {
	return NewApiErr(http.StatusBadRequest, message)
}

func NewNotFound(resource string) *ApiErr {
	return &ApiErr{http.StatusNotFound, fmt.Errorf("%s %w", resource, NotFound)}
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

func AlreadyExists(resource string) *ApiErr {
	return NewApiErr(http.StatusBadRequest, resource+" already exists")
}
