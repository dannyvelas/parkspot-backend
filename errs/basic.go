package errs

import (
	"fmt"
	"net/http"
)

var (
	Unauthorized = &ApiErr{http.StatusUnauthorized, "unauthorized"}
	NotFound     = &ApiErr{http.StatusNotFound, "not found"}
)

func EmptyFields(fields string) *ApiErr {
	return &ApiErr{http.StatusBadRequest, "One or more missing fields: " + fields}
}

func InvalidFields(fields string) *ApiErr {
	return &ApiErr{http.StatusBadRequest, "One or more invalid fields: " + fields}
}

func Malformed(payload string) *ApiErr {
	return &ApiErr{http.StatusBadRequest, payload + " malformed"}
}

func AlreadyExists(resource string) *ApiErr {
	return &ApiErr{http.StatusBadRequest, resource + " already exists"}
}

func Internalf(message string, args ...any) *ApiErr {
	return &ApiErr{http.StatusInternalServerError, fmt.Sprintf(message, args...)}
}
