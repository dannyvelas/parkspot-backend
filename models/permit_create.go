package models

type CreatePermit struct {
	ResidentId      string
	StartDate       int64
	EndDate         int64
	ExceptionReason string
}
