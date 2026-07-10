package errs

import (
	"errors"
)

type APIErr struct {
	StatusCode int
	err        error
}

func NewAPIErr(statusCode int, message string) *APIErr {
	return &APIErr{
		StatusCode: statusCode,
		err:        errors.New(message),
	}
}

// Error implements error interface
func (e *APIErr) Error() string {
	return e.err.Error()
}

func (e *APIErr) Unwrap() error {
	return e.err
}
