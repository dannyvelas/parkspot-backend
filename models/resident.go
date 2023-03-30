package models

import (
	"github.com/dannyvelas/lasvistas_api/errs"
	"regexp"
	"strings"
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
	TokenVersion       int    `json:"-"`
}

func NewResident(
	id string,
	firstName string,
	lastName string,
	phone string,
	email string,
	password string,
	unlimDays *bool,
	amtParkingDaysUsed *int,
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

func (m Resident) GetPassword() string {
	return m.Password
}

func (m Resident) AsUser() User {
	return NewUser(m.ID, m.FirstName, m.LastName, m.Email, ResidentRole, m.TokenVersion)
}

func (m Resident) ValidateEdit() *errs.ApiErr {
	if m.FirstName == "" &&
		m.LastName == "" &&
		m.Phone == "" &&
		m.Email == "" &&
		m.UnlimDays == nil &&
		m.AmtParkingDaysUsed == nil {
		return errs.EmptyFields("firstName, lastName, phone, email, unlimDays, amtParkingDaysUsed")
	}

	errors := []string{}
	if m.Phone != "" &&
		!regexp.MustCompile("^\\d{1,20}$").MatchString(m.Phone) {
		errors = append(errors, "phone number must be only digits, no longer than 20")
	}
	if m.Email != "" && !strings.Contains(m.Email, "@") {
		errors = append(errors, "email must have an '@'")
	}
	if m.AmtParkingDaysUsed != nil && *m.AmtParkingDaysUsed < 0 {
		errors = append(errors, "amtParkingDaysUsed field must be 0 or positive.")
	}

	if len(errors) > 0 {
		return errs.InvalidFields(strings.Join(errors, ". "))
	}

	return nil
}

func (m Resident) ValidateCreation() *errs.ApiErr {
	if err := m.emptyFields(); err != nil {
		return err
	}

	if err := m.invalidFields(); err != nil {
		return err
	}

	return nil
}

func (m Resident) emptyFields() *errs.ApiErr {
	emptyFields := []string{}

	if m.ID == "" {
		emptyFields = append(emptyFields, "residentID")
	}
	if m.FirstName == "" {
		emptyFields = append(emptyFields, "firstName")
	}
	if m.LastName == "" {
		emptyFields = append(emptyFields, "lastName")
	}
	if m.Phone == "" {
		emptyFields = append(emptyFields, "phone")
	}
	if m.Email == "" {
		emptyFields = append(emptyFields, "email")
	}
	if m.Password == "" {
		emptyFields = append(emptyFields, "password")
	}

	if len(emptyFields) > 0 {
		return errs.EmptyFields(strings.Join(emptyFields, ", "))
	}

	return nil
}

func (m Resident) invalidFields() *errs.ApiErr {
	errors := []string{}

	if err := IsResidentID(m.ID); err != nil {
		errors = append(errors, err.Error())
	}
	if !regexp.MustCompile("^\\d{1,20}$").MatchString(m.Phone) {
		errors = append(errors, "phone number must be only digits, no longer than 20")
	}
	if !strings.Contains(m.Email, "@") {
		errors = append(errors, "email must have an '@'")
	}

	if len(errors) > 0 {
		return errs.InvalidFields(strings.Join(errors, ". "))
	}

	return nil
}

func IsResidentID(s string) error {
	if !regexp.MustCompile("^(B|T)\\d{7}$").MatchString(s) {
		return errs.InvalidResID
	}

	return nil
}
