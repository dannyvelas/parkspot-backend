package models

import (
	"github.com/dannyvelas/parkspot-backend/errs"
	"regexp"
)

type Resident struct {
	ID                 string `json:"id"`
	FirstName          string `json:"firstName"`
	LastName           string `json:"lastName"`
	Phone              string `json:"phone"`
	Email              string `json:"email"`
	Password           string `json:"password"`
	UnlimDays          *bool  `json:"unlimDays"`
	AmtParkingDaysUsed *int   `json:"amtParkingDaysUsed"`
	TokenVersion       *int   `json:"-"`
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
		UnlimDays:          &unlimDays,
		AmtParkingDaysUsed: &amtParkingDaysUsed,
		TokenVersion:       &tokenVersion,
	}
}

func (m Resident) GetPassword() string {
	return m.Password
}

func (m Resident) AsUser() User {
	return NewUser(m.ID, m.FirstName, m.LastName, m.Email, ResidentRole, *m.TokenVersion)
}

func IsResidentID(s string) error {
	if !regexp.MustCompile("^(B|T)\\d{7}$").MatchString(s) {
		return errs.InvalidResID
	}

	return nil
}
