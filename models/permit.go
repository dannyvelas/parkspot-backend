package models

import (
	"fmt"
	"strings"
	"time"
)

type Permit struct {
	ID              int       `json:"id"`
	ResidentID      string    `json:"residentID"`
	CarID           string    `json:"carID"`
	LicensePlate    string    `json:"licensePlate"`
	Color           string    `json:"color"`
	Make            string    `json:"make"`
	Model           string    `json:"model"`
	StartDate       time.Time `json:"startDate"`
	EndDate         time.Time `json:"endDate"`
	RequestTS       int64     `json:"requestTS"` // int64: type used by time package for unix time
	AffectsDays     bool      `json:"affectsDays"`
	ExceptionReason string    `json:"exceptionReason"`
}

func NewPermit(
	id int,
	residentID string,
	carID string,
	licensePlate string,
	color string,
	make string,
	model string,
	startDate time.Time,
	endDate time.Time,
	requestTS int64,
	affectsDays bool,
	exceptionReason string,
) Permit {
	return Permit{
		ID:              id,
		ResidentID:      residentID,
		CarID:           carID,
		LicensePlate:    licensePlate,
		Color:           color,
		Make:            make,
		Model:           model,
		StartDate:       startDate,
		EndDate:         endDate,
		RequestTS:       requestTS,
		AffectsDays:     affectsDays,
		ExceptionReason: exceptionReason,
	}
}

func (self Permit) Equal(other Permit) bool {
	if self.ID != other.ID {
		return false
	} else if self.ResidentID != other.ResidentID {
		return false
	} else if self.CarID != other.CarID {
		return false
	} else if self.LicensePlate != other.LicensePlate {
		return false
	} else if self.Color != other.Color {
		return false
	} else if self.Make != other.Make {
		return false
	} else if self.Model != other.Model {
		return false
	} else if !self.StartDate.Equal(other.StartDate) {
		return false
	} else if !self.EndDate.Equal(other.EndDate) {
		return false
	} else if self.RequestTS != other.RequestTS {
		return false
	} else if self.AffectsDays != other.AffectsDays {
		return false
	} else if self.ExceptionReason != other.ExceptionReason {
		return false
	}

	return true
}

func (m Permit) ValidateCreation() error {
	if err := m.emptyFields(); err != nil {
		return err
	}

	if err := m.invalidFields(); err != nil {
		return err
	}

	return nil
}

func (m Permit) emptyFields() error {
	emptyFields := []string{}

	if m.ResidentID == "" {
		emptyFields = append(emptyFields, "residentID")
	}
	if m.CarID == "" {
		emptyFields = append(emptyFields, "carID")
	}
	if m.StartDate.IsZero() {
		emptyFields = append(emptyFields, "startDate")
	}
	if m.EndDate.IsZero() {
		emptyFields = append(emptyFields, "endDate")
	}

	if len(emptyFields) > 0 {
		return fmt.Errorf("%w: %v", ErrEmptyFields, strings.Join(emptyFields, ", "))
	}

	return nil
}

func (m Permit) invalidFields() error {
	errors := []string{}

	if m.ResidentID[0] == 'P' {
		errors = append(errors, "Accounts with a ResidentID starting with 'P' are not allowed to request permits")
	} else if err := IsResidentID(m.ResidentID); err != nil {
		errors = append(errors, err.Error())
	}

	if m.StartDate.After(m.EndDate) {
		errors = append(errors, "startDate cannot be after endDate")
	}

	if m.StartDate.Equal(m.EndDate) {
		errors = append(errors, "startDate cannot be equal to endDate")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%w: %v", ErrInvalidFields, strings.Join(errors, ". "))
	}

	return nil
}
