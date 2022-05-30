package storage

import (
	"fmt"
	"strings"
)

type permitFilter int

const (
	All permitFilter = iota
	Active
	Expired
	Exceptions
)

func NewPermitFilter(s string) (permitFilter, error) {
	switch strings.ToLower(s) {
	case "", "all":
		return All, nil
	case "active":
		return Active, nil
	case "expired":
		return Expired, nil
	case "exceptions":
		return Exceptions, nil
	default:
		return All, fmt.Errorf("Invalid permitFilter: %w", ErrInvalidArg)
	}
}
