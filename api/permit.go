package api

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/models"
	"regexp"
	"strings"
	"time"
)

type createPermitReq struct {
	residentId      string       `json:"residentId"`
	createCarReq    createCarReq `json:"car"`
	startDate       time.Time    `json:"startDate"`
	endDate         time.Time    `json:"endDate"`
	requestTS       int64        `json:"requestTS"`
	affectsDays     bool         `json:"affectsDays"`
	exceptionReason *string      `json:"exceptionReason"`
}

func (createPermitReq createPermitReq) emptyFields() error {
	emptyFields := []string{}

	if createPermitReq.residentId == "" {
		emptyFields = append(emptyFields, "residentId")
	}
	if createPermitReq.startDate.IsZero() {
		emptyFields = append(emptyFields, "startDate")
	}
	if createPermitReq.endDate.IsZero() {
		emptyFields = append(emptyFields, "endDate")
	}
	if createPermitReq.requestTS == 0 {
		emptyFields = append(emptyFields, "requestTS")
	}
	if createPermitReq.affectsDays == false {
		// this is okay so do nothing
	}
	if createPermitReq.exceptionReason == nil {
		// this is okay so do nothing
	} else if *createPermitReq.exceptionReason == "" {
		emptyFields = append(emptyFields, "exceptionReason")
	}

	if len(emptyFields) > 0 {
		return fmt.Errorf("%w: %v", errEmptyFields, strings.Join(emptyFields, ", "))
	}

	return nil
}

func (createPermitReq createPermitReq) invalidFields() error {
	errors := []string{}

	if createPermitReq.residentId[0] == 'P' {
		errors = append(errors, "Accounts with a ResidentId starting with 'P' are not allowed to request permits")
	} else if !regexp.MustCompile("^(B|T)\\d{7}$").MatchString(createPermitReq.residentId) {
		errors = append(errors, "residentId must start be a 'B' or a 'T', followed by 7 numbers")
	}

	if err := createPermitReq.createCarReq.validate(); err != nil {
		errors = append(errors, err.Error())
	}

	if createPermitReq.startDate.After(createPermitReq.endDate) {
		errors = append(errors, "startDate cannot be after endDate")
	}

	if createPermitReq.startDate.Equal(createPermitReq.endDate) {
		errors = append(errors, "startDate cannot be equal to endDate")
	}

	if createPermitReq.requestTS > time.Now().Unix() {
		errors = append(errors, "requestTS cannot be in the future")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%w: %v", errInvalidFields, strings.Join(errors, ". "))
	}

	return nil
}

func (createPermitReq createPermitReq) validate() error {
	if err := createPermitReq.emptyFields(); err != nil {
		return err
	}

	if err := createPermitReq.invalidFields(); err != nil {
		return err
	}

	return nil
}

func (createPermitReq createPermitReq) toNewPermitArgs(carId string) models.NewPermitArgs {
	return models.NewNewPermitArgs(
		createPermitReq.residentId,
		carId,
		createPermitReq.startDate,
		createPermitReq.endDate,
		createPermitReq.requestTS,
		createPermitReq.affectsDays,
		createPermitReq.exceptionReason,
	)
}
