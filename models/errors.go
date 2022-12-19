package models

import (
	"errors"
)

var (
	ErrEmptyFields   = errors.New("One or more missing fields")
	ErrInvalidFields = errors.New("One or more invalid fields")
)
