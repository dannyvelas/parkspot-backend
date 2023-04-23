package models

type Status int

const (
	AnyStatus Status = iota
	ActiveStatus
	ExpiredStatus
	ExceptionStatus
)
