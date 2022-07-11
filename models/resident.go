package models

import (
	"errors"
	"regexp"
)

type Resident struct {
	Id                 string `json:"id"`
	FirstName          string `json:"firstName"`
	LastName           string `json:"lastName"`
	Phone              string `json:"phone"`
	Email              string `json:"email"`
	Password           string `json:"-"`
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

func IsResidentId(s string) error {
	if !regexp.MustCompile("^(B|T)\\d{7}$").MatchString(s) {
		return errors.New("residentId must start be a 'B' or a 'T', followed by 7 numbers")
	}

	return nil
}
