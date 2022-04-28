package api

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/models"
	"regexp"
	"strings"
	"time"
)

type createPermitReq struct {
	ResidentId      string       `json:"residentId"`
	CreateCarReq    createCarReq `json:"car"`
	StartDate       time.Time    `json:"startDate"`
	EndDate         time.Time    `json:"endDate"`
	RequestTS       int64        `json:"requestTS"`
	AffectsDays     bool         `json:"affectsDays"`
	ExceptionReason *string      `json:"exceptionReason"`
}

func (createPermitReq createPermitReq) emptyFields() error {
	emptyFields := []string{}

	if createPermitReq.ResidentId == "" {
		emptyFields = append(emptyFields, "residentId")
	}
	if createPermitReq.StartDate.IsZero() {
		emptyFields = append(emptyFields, "startDate")
	}
	if createPermitReq.EndDate.IsZero() {
		emptyFields = append(emptyFields, "endDate")
	}
	if createPermitReq.RequestTS == 0 {
		emptyFields = append(emptyFields, "requestTS")
	}
	if createPermitReq.AffectsDays == false {
		// this is okay so do nothing
	}
	if createPermitReq.ExceptionReason == nil {
		// this is okay so do nothing
	} else if *createPermitReq.ExceptionReason == "" {
		emptyFields = append(emptyFields, "exceptionReason")
	}

	if len(emptyFields) > 0 {
		return fmt.Errorf("%w: %v", errEmptyFields, strings.Join(emptyFields, ", "))
	}

	return nil
}

func (createPermitReq createPermitReq) invalidFields() error {
	errors := []string{}

	if createPermitReq.ResidentId[0] == 'P' {
		errors = append(errors, "Accounts with a ResidentId starting with 'P' are not allowed to request permits")
	} else if !regexp.MustCompile("^(B|T)\\d{7}$").MatchString(createPermitReq.ResidentId) {
		errors = append(errors, "residentId must start be a 'B' or a 'T', followed by 7 numbers")
	}

	if createPermitReq.StartDate.After(createPermitReq.EndDate) {
		errors = append(errors, "startDate cannot be after endDate")
	}

	if createPermitReq.StartDate.Equal(createPermitReq.EndDate) {
		errors = append(errors, "startDate cannot be equal to endDate")
	}

	if createPermitReq.RequestTS > time.Now().Unix() {
		errors = append(errors, "requestTS cannot be in the future")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%w: %v", errInvalidFields, strings.Join(errors, ". "))
	}

	return nil
}

func (createPermitReq createPermitReq) toModels() (models.CreatePermit, error) {
	if err := createPermitReq.emptyFields(); err != nil {
		return models.CreatePermit{}, err
	}

	if err := createPermitReq.invalidFields(); err != nil {
		return models.CreatePermit{}, err
	}

	createCar, err := createPermitReq.CreateCarReq.toModels()
	if err != nil {
		return models.CreatePermit{}, fmt.Errorf("%w: %v", errInvalidFields, "car")
	}

	return models.NewCreatePermit(
		createPermitReq.ResidentId,
		createCar,
		createPermitReq.StartDate,
		createPermitReq.EndDate,
		createPermitReq.RequestTS,
		createPermitReq.AffectsDays,
		createPermitReq.ExceptionReason,
	), nil
}
