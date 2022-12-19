package models

type CreatePermit struct {
	ResidentId      string
	CarId           string
	StartDate       int64
	EndDate         int64
	AffectsDays     bool
	ExceptionReason string
}
