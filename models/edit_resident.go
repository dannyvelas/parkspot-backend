package models

type EditResident struct {
	FirstName          string `json:"firstName"`
	LastName           string `json:"lastName"`
	Phone              string `json:"phone"`
	Email              string `json:"email"`
	Password           string `json:"password"`
	UnlimDays          *bool  `json:"unlimDays"`
	AmtParkingDaysUsed *int   `json:"amtParkingDaysUsed"`
	TokenVersion       int    `json:"-"`
}
