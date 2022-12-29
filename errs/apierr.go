package errs

import (
	"errors"
)

type ApiErr struct {
	StatusCode int
	err        error
}

func NewApiErr(statusCode int, message string) *ApiErr {
	return &ApiErr{
		StatusCode: statusCode,
		err:        errors.New(message),
	}
}

// implements error interface
func (e *ApiErr) Error() string {
	return e.err.Error()
}

func (e *ApiErr) Unwrap() error {
	return e.err
}
