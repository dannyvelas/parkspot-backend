package models

type Permit struct {
	Id            int
	ResidentId    string
	LicensePlate  string
	ColorAndModel string
	StartDate     int
	EndDate       int
	RequestDate   int
	AffectsDays   bool
}
