package models

type Resident struct {
	Id                 string
	FirstName          string
	LastName           string
	Phone              string
	Email              string
	Password           string
	UnlimDays          bool
	AmtParkingDaysUsed int
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
