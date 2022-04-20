package models

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"strings"
	"time"
)

type Permit struct {
	Id int `json:"id"`
	PermitFields
}

func (self Permit) Equal(other Permit) bool {
	if self.Id != other.Id {
		return false
	} else if self.ResidentId != other.ResidentId {
		return false
	} else if !cmp.Equal(self.Car, other.Car) {
		return false
	} else if !self.StartDate.Equal(other.StartDate) {
		return false
	} else if !self.EndDate.Equal(other.EndDate) {
		return false
	} else if self.RequestTS != other.RequestTS {
		return false
	} else if self.AffectsDays != other.AffectsDays {
		return false
	}

	return true
}

type PermitFields struct {
	ResidentId  string    `json:"residentId"`
	Car         Car       `json:"car"`
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate"`
	RequestTS   int64     `json:"requestTS"`
	AffectsDays bool      `json:"affectsDays"`
}

func (permitFields PermitFields) emptyFields() error {
	emptyFields := []string{}

	if permitFields.ResidentId == "" {
		emptyFields = append(emptyFields, "ResidentId")
	} else if permitFields.StartDate.IsZero() {
		emptyFields = append(emptyFields, "StartDate")
	} else if permitFields.EndDate.IsZero() {
		emptyFields = append(emptyFields, "EndDate")
	} else if permitFields.RequestTS == 0 {
		emptyFields = append(emptyFields, "RequestTS")
	} else if permitFields.AffectsDays == false {
		// this is okay so do nothing
	}

	if len(emptyFields) > 0 {
		return fmt.Errorf("%w: %v", ErrEmptyFields, strings.Join(emptyFields, ", "))
	}

	return nil
}

func (permitFields PermitFields) invalidFields() error {
	errors := []string{}

	if permitFields.ResidentId[0] == 'P' {
		errors = append(errors, "Accounts with a ResidentId starting with 'P' are not allowed to request permits")
	}

	if err := permitFields.Car.Validate(); err != nil {
		errors = append(errors, fmt.Sprintf("invalid car: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("%v", strings.Join(errors, ". "))
	}

	return nil
}

func (permitFields PermitFields) Validate() error {
	if err := permitFields.emptyFields(); err != nil {
		return err
	}

	if err := permitFields.invalidFields(); err != nil {
		return err
	}

	return nil
}
