package models

type Permit struct {
	Id          int
	ResidentId  string
	Car         Car
	StartDate   int
	EndDate     int
	RequestDate int
	AffectsDays bool
}
