package api

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/models"
	"regexp"
	"strings"
)

type newResidentReq struct {
	ResidentID string `json:"residentID"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	UnlimDays  bool   `json:"unlimDays"`
}

func (newResidentReq newResidentReq) emptyFields() error {
	emptyFields := []string{}

	if newResidentReq.ResidentID == "" {
		emptyFields = append(emptyFields, "residentID")
	}
	if newResidentReq.FirstName == "" {
		emptyFields = append(emptyFields, "firstName")
	}
	if newResidentReq.LastName == "" {
		emptyFields = append(emptyFields, "lastName")
	}
	if newResidentReq.Phone == "" {
		emptyFields = append(emptyFields, "phone")
	}
	if newResidentReq.Email == "" {
		emptyFields = append(emptyFields, "email")
	}
	if newResidentReq.Password == "" {
		emptyFields = append(emptyFields, "password")
	}
	if newResidentReq.UnlimDays == false {
		// noop: this is okay as this is an optional field
	}

	if len(emptyFields) > 0 {
		return fmt.Errorf("%w: %v", errEmptyFields, strings.Join(emptyFields, ", "))
	}

	return nil
}

func (newResidentReq newResidentReq) invalidFields() error {
	errors := []string{}

	if err := models.IsResidentID(newResidentReq.ResidentID); err != nil {
		errors = append(errors, err.Error())
	}
	if !regexp.MustCompile("^\\d{1,20}$").MatchString(newResidentReq.Phone) {
		errors = append(errors, "phone number must be only digits, no longer than 20")
	}
	if !strings.Contains(newResidentReq.Email, "@") {
		errors = append(errors, "email must have an '@'")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%w: %v", errInvalidFields, strings.Join(errors, ". "))
	}

	return nil
}

func (newResidentReq newResidentReq) validate() error {
	if err := newResidentReq.emptyFields(); err != nil {
		return err
	}

	if err := newResidentReq.invalidFields(); err != nil {
		return err
	}

	return nil
}
