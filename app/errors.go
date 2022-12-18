package app

import (
	"errors"
	"fmt"
)

var (
	ErrUnauthorized  = errors.New("unauthorized")
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

func newErrAlreadyExists(resource string) error {
	return fmt.Errorf("error: %s %w", resource, ErrAlreadyExists)
}
