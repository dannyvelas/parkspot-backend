package models

type Resident struct {
	Id                 string `json:"id"`
	FirstName          string `json:"firstName"`
	LastName           string `json:"lastName"`
	Phone              string `json:"phone"`
	Email              string `json:"email"`
	Password           string `json:"password"`
	UnlimDays          bool   `json:"unlimDays"`
	AmtParkingDaysUsed int    `json:"amtParkingDaysUsed"`
}

func NewResident(
	id string,
	firstName string,
	lastName string,
	phone string,
	email string,
	password string,
	unlimDays bool,
	amtParkingDaysUsed int,
) Resident {
	return Resident{
		Id:                 id,
		FirstName:          firstName,
		LastName:           lastName,
		Phone:              phone,
		Email:              email,
		Password:           password,
		UnlimDays:          unlimDays,
		AmtParkingDaysUsed: amtParkingDaysUsed,
	}
}
