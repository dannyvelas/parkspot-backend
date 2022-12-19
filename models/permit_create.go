package models

import (
	"fmt"
	"strings"
	"time"
)

type CreatePermit struct {
	ResidentId      string    `json:"residentId"`
	Car             CreateCar `json:"car"`
	StartDate       time.Time `json:"startDate"`
	EndDate         time.Time `json:"endDate"`
	ExceptionReason string    `json:"exceptionReason"`
}

func (createPermit CreatePermit) emptyFields() error {
	emptyFields := []string{}

	if createPermit.ResidentId == "" {
		emptyFields = append(emptyFields, "residentId")
	}
	if createPermit.StartDate.IsZero() {
		emptyFields = append(emptyFields, "startDate")
	}
	if createPermit.EndDate.IsZero() {
		emptyFields = append(emptyFields, "endDate")
	}
	if createPermit.ExceptionReason == "" {
		// this is okay so do nothing
	}

	if len(emptyFields) > 0 {
		return fmt.Errorf("%w: %v", ErrEmptyFields, strings.Join(emptyFields, ", "))
	}

	return nil
}

func (createPermit CreatePermit) invalidFields() error {
	errors := []string{}

	if createPermit.ResidentId[0] == 'P' {
		errors = append(errors, "Accounts with a ResidentId starting with 'P' are not allowed to request permits")
	} else if err := IsResidentId(createPermit.ResidentId); err != nil {
		errors = append(errors, err.Error())
	}

	if err := createPermit.Car.Validate(); err != nil {
		errors = append(errors, err.Error())
	}

	if createPermit.StartDate.After(createPermit.EndDate) {
		errors = append(errors, "startDate cannot be after endDate")
	}

	if createPermit.StartDate.Equal(createPermit.EndDate) {
		errors = append(errors, "startDate cannot be equal to endDate")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%w: %v", ErrInvalidFields, strings.Join(errors, ". "))
	}

	return nil
}

func (createPermit CreatePermit) Validate() error {
	if err := createPermit.emptyFields(); err != nil {
		return err
	}

	if err := createPermit.invalidFields(); err != nil {
		return err
	}

	return nil
}
