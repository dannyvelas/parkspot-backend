package models

type CreatePermit struct {
	ResidentID      string
	StartDate       int64
	EndDate         int64
	ExceptionReason string
}
