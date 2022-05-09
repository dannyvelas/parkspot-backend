package api

import (
	"errors"
	"strings"
)

type Role uint8

const (
	UndefinedRole Role = iota
	AdminRole
	ResidentRole
)

func NewRole(value string) (Role, error) {
	switch strings.ToLower(value) {
	case "admin":
		return AdminRole, nil
	case "resident":
		return ResidentRole, nil
	default:
		return UndefinedRole, errors.New("Invalid auth role of " + value)
	}
}

func String(role Role) string {
	switch role {
	case AdminRole:
		return "Admin"
	case ResidentRole:
		return "Resident"
	default:
		return ""
	}
}
