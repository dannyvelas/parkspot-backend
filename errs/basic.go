package errs

import (
	"net/http"
)

var (
	Unauthorized = &ApiErr{http.StatusUnauthorized, "unauthorized"}
)

func BadRequest(message string) *ApiErr {
	return &ApiErr{http.StatusBadRequest, message}
}

func NotFound(resource string) *ApiErr {
	return &ApiErr{http.StatusNotFound, resource + " not found"}
}

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
