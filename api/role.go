package api

type Role uint8

const (
	UndefinedRole Role = iota
	AdminRole
	ResidentRole
)
