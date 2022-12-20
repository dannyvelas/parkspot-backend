package models

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Resident struct {
	ID                 string `json:"id"`
	FirstName          string `json:"firstName"`
	LastName           string `json:"lastName"`
	Phone              string `json:"phone"`
	Email              string `json:"email"`
	Password           string `json:"-"`
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

func (m Resident) ValidateEdit() error {
	if m.FirstName == "" &&
		m.LastName == "" &&
		m.Phone == "" &&
		m.Email == "" &&
		m.UnlimDays == nil &&
		m.AmtParkingDaysUsed == nil {
		return fmt.Errorf("%w: %v", ErrEmptyFields, "all edit fields cannot be empty")
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
		errors = append(errors, "amountParkingDaysUsed field must be 0 or positive.")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%w: %v", ErrInvalidFields, strings.Join(errors, ". "))
	}

	return nil
}

func (m Resident) ValidateCreation() error {
	if err := m.emptyFields(); err != nil {
		return err
	}

	if err := m.invalidFields(); err != nil {
		return err
	}

	return nil
}

func (m Resident) emptyFields() error {
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
	if m.UnlimDays == nil {
		// this is okay as this is an optional field
		*m.UnlimDays = false
	}

	if len(emptyFields) > 0 {
		return fmt.Errorf("%w: %v", ErrEmptyFields, strings.Join(emptyFields, ", "))
	}

	return nil
}

func (m Resident) invalidFields() error {
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
		return fmt.Errorf("%w: %v", ErrInvalidFields, strings.Join(errors, ". "))
	}

	return nil
}

func IsResidentID(s string) error {
	if !regexp.MustCompile("^(B|T)\\d{7}$").MatchString(s) {
		return errors.New("residentID must start be a 'B' or a 'T', followed by 7 numbers")
	}

	return nil
}
