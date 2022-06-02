package models

import (
	"fmt"
	"strings"
)

type PermitFilter int

const (
	AllPermits PermitFilter = iota
	ActivePermits
	ExpiredPermits
	ExceptionPermits
)

func NewPermitFilter(s string) (PermitFilter, error) {
	switch strings.ToLower(s) {
	case "", "all":
		return AllPermits, nil
	case "active":
		return ActivePermits, nil
	case "expired":
		return ExpiredPermits, nil
	case "exceptions":
		return ExceptionPermits, nil
	default:
		return AllPermits, fmt.Errorf("Invalid permitFilter")
	}
}
