package api

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/models"
	"regexp"
	"strings"
	"time"
)

type newPermitReq struct {
	ResidentId      string    `json:"residentId"`
	NewCarReq       newCarReq `json:"car"`
	StartDate       time.Time `json:"startDate"`
	EndDate         time.Time `json:"endDate"`
	ExceptionReason string    `json:"exceptionReason"`
}

func (newPermitReq newPermitReq) emptyFields() error {
	emptyFields := []string{}

	if newPermitReq.ResidentId == "" {
		emptyFields = append(emptyFields, "residentId")
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

	if newPermitReq.ResidentId[0] == 'P' {
		errors = append(errors, "Accounts with a ResidentId starting with 'P' are not allowed to request permits")
	} else if !regexp.MustCompile("^(B|T)\\d{7}$").MatchString(newPermitReq.ResidentId) {
		errors = append(errors, "residentId must start be a 'B' or a 'T', followed by 7 numbers")
	}

	if err := newPermitReq.NewCarReq.validate(); err != nil {
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

func (newPermitReq newPermitReq) toNewPermitArgs(carId string, affectsDays bool) models.NewPermitArgs {
	return models.NewNewPermitArgs(
		newPermitReq.ResidentId,
		carId,
		newPermitReq.StartDate,
		newPermitReq.EndDate,
		affectsDays,
		newPermitReq.ExceptionReason,
	)
}
