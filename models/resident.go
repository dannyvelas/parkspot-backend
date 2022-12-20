package models

import (
	"errors"
	"regexp"
)

type Resident struct {
	ID                 string `json:"id"`
	FirstName          string `json:"firstName"`
	LastName           string `json:"lastName"`
	Phone              string `json:"phone"`
	Email              string `json:"email"`
	Password           string `json:"-"`
	UnlimDays          bool   `json:"unlimDays"`
	AmtParkingDaysUsed int    `json:"amtParkingDaysUsed"`
	TokenVersion       int    `json:"-"`
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
	tokenVersion int,
) Resident {
	return Resident{
		ID:                 id,
		FirstName:          firstName,
		LastName:           lastName,
		Phone:              phone,
		Email:              email,
		Password:           password,
		UnlimDays:          unlimDays,
		AmtParkingDaysUsed: amtParkingDaysUsed,
		TokenVersion:       tokenVersion,
	}
}

func (r Resident) GetPassword() string {
	return r.Password
}

func (r Resident) AsUser() User {
	return newUser(r.ID, r.FirstName, r.LastName, r.Email, ResidentRole, r.TokenVersion)
}

func IsResidentID(s string) error {
	if !regexp.MustCompile("^(B|T)\\d{7}$").MatchString(s) {
		return errors.New("residentID must start be a 'B' or a 'T', followed by 7 numbers")
	}

	return nil
}
