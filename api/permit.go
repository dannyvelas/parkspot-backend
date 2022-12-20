package api

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/models"
	"strings"
	"time"
)

type newPermitReq struct {
	ResidentID      string     `json:"residentID"`
	Car             models.Car `json:"car"`
	StartDate       time.Time  `json:"startDate"`
	EndDate         time.Time  `json:"endDate"`
	ExceptionReason string     `json:"exceptionReason"`
}

func (newPermitReq newPermitReq) emptyFields() error {
	emptyFields := []string{}

	if newPermitReq.ResidentID == "" {
		emptyFields = append(emptyFields, "residentID")
	}
	if newPermitReq.StartDate.IsZero() {
		emptyFields = append(emptyFields, "startDate")
	}
	if newPermitReq.EndDate.IsZero() {
		emptyFields = append(emptyFields, "endDate")
	}
	if newPermitReq.ExceptionReason == "" {
		// this is okay so do nothing
	}

	if len(emptyFields) > 0 {
		return fmt.Errorf("%w: %v", errEmptyFields, strings.Join(emptyFields, ", "))
	}

	return nil
}

func (newPermitReq newPermitReq) invalidFields() error {
	errors := []string{}

	if newPermitReq.ResidentID[0] == 'P' {
		errors = append(errors, "Accounts with a ResidentID starting with 'P' are not allowed to request permits")
	} else if err := models.IsResidentID(newPermitReq.ResidentID); err != nil {
		errors = append(errors, err.Error())
	}

	if err := newPermitReq.Car.ValidateCreation(); err != nil {
		errors = append(errors, err.Error())
	}

	if newPermitReq.StartDate.After(newPermitReq.EndDate) {
		errors = append(errors, "startDate cannot be after endDate")
	}

	if newPermitReq.StartDate.Equal(newPermitReq.EndDate) {
		errors = append(errors, "startDate cannot be equal to endDate")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%w: %v", errInvalidFields, strings.Join(errors, ". "))
	}

	return nil
}

func (newPermitReq newPermitReq) validate() error {
	if err := newPermitReq.emptyFields(); err != nil {
		return err
	}

	if err := newPermitReq.invalidFields(); err != nil {
		return err
	}

	return nil
}
