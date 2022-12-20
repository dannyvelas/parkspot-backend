package models

import (
	"fmt"
	"regexp"
	"strings"
)

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

func (m EditResident) Validate() error {
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
