package errs

import (
	"errors"
	"fmt"
	"net/http"
)

var ErrCarWithLPAlreadyExists = errors.New("a car with this licensePlate")

func NewErrCarWithLPAlreadyExists(licensePlate string) *APIErr {
	return &APIErr{
		http.StatusBadRequest,
		fmt.Errorf("%w %s %w", ErrCarWithLPAlreadyExists, licensePlate, AlreadyExists),
	}
}
