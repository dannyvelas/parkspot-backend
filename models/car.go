package models

import (
	"github.com/dannyvelas/lasvistas_api/errs"
	"regexp"
	"strings"
)

type Car struct {
	ID                 string `json:"id"`
	LicensePlate       string `json:"licensePlate"`
	Color              string `json:"color"`
	Make               string `json:"make"`
	Model              string `json:"model"`
	AmtParkingDaysUsed int    `json:"amtParkingDaysUsed"`
}

func NewCar(id string, licensePlate string, color string, make string, model string, amtParkingDaysUsed int) Car {
	return Car{
		ID:                 id,
		LicensePlate:       licensePlate,
		Color:              color,
		Make:               make,
		Model:              model,
		AmtParkingDaysUsed: amtParkingDaysUsed,
	}
}

func (self Car) Equal(other Car) bool {
	if self.ID != other.ID {
		return false
	} else if self.LicensePlate != other.LicensePlate {
		return false
	} else if self.Color != other.Color {
		return false
	} else if self.Make != other.Make {
		return false
	} else if self.Model != other.Model {
		return false
	} else if self.AmtParkingDaysUsed != other.AmtParkingDaysUsed {
		return false
	}

	return true
}

func (m Car) ValidateEdit() *errs.ApiErr {
	if m.Color == "" && m.Make == "" && m.Model == "" {
		return errs.EmptyFields("color, make, model")
	}

	if errors := getColorMakeModelErrors(m.Color, m.Make, m.Model); len(errors) != 0 {
		return errs.InvalidFields(strings.Join(errors, ". "))
	}

	return nil
}

func (m Car) ValidateCreation() *errs.ApiErr {
	if err := m.emptyFields(); err != nil {
		return err
	}

	if err := m.invalidFields(); err != nil {
		return err
	}

	return nil
}

func (m Car) emptyFields() *errs.ApiErr {
	emptyFields := []string{}

	if m.LicensePlate == "" {
		emptyFields = append(emptyFields, "licensePlate")
	}
	if m.Color == "" {
		emptyFields = append(emptyFields, "color")
	}
	if m.Make == "" {
		emptyFields = append(emptyFields, "make")
	}
	if m.Model == "" {
		emptyFields = append(emptyFields, "model")
	}

	if len(emptyFields) > 0 {
		return errs.EmptyFields(strings.Join(emptyFields, ", "))
	}

	return nil
}

func (m Car) invalidFields() *errs.ApiErr {
	errors := []string{}

	if !regexp.MustCompile("^[A-Za-z0-9]+$").MatchString(m.LicensePlate) {
		errors = append(errors, "licensePlate can only be letters or numbers")
	}
	if len(m.LicensePlate) > 8 {
		errors = append(errors, "licensePlate can be maximum 8 characters")
	}
	if colorMakeModelErrors := getColorMakeModelErrors(m.Color, m.Make, m.Model); len(errors) != 0 {
		errors = append(errors, colorMakeModelErrors...)
	}

	if len(errors) > 0 {
		return errs.InvalidFields(strings.Join(errors, ". "))
	}

	return nil
}
