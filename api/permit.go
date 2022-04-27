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
		emptyFields = append(emptyFields, "ResidentId")
	}
	if createPermitReq.StartDate.IsZero() {
		emptyFields = append(emptyFields, "StartDate")
	}
	if createPermitReq.EndDate.IsZero() {
		emptyFields = append(emptyFields, "EndDate")
	}
	if createPermitReq.RequestTS == 0 {
		emptyFields = append(emptyFields, "RequestTS")
	}
	if createPermitReq.AffectsDays == false {
		// this is okay so do nothing
	}
	if createPermitReq.ExceptionReason == nil {
		// this is okay so do nothing
	} else if *createPermitReq.ExceptionReason == "" {
		emptyFields = append(emptyFields, "ExceptionReason")
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
		errors = append(errors, "ResidentId must start be a 'B' or a 'T', followed by 7 numbers")
	}

	if createPermitReq.RequestTS > time.Now().Unix() {
		errors = append(errors, "RequestTS cannot be in the future")
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

	return models.CreatePermit{
		ResidentId:  createPermitReq.ResidentId,
		CreateCar:   createCar,
		StartDate:   createPermitReq.StartDate,
		EndDate:     createPermitReq.EndDate,
		RequestTS:   createPermitReq.RequestTS,
		AffectsDays: createPermitReq.AffectsDays,
	}, nil
}
