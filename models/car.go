package models

import (
	"fmt"
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

func (m Car) ValidateEdit() error {
	if m.Color == "" && m.Make == "" && m.Model == "" {
		return fmt.Errorf("%w: %v", ErrEmptyFields, "all edit fields cannot be empty")
	}

	if errors := getColorMakeModelErrors(m.Color, m.Make, m.Model); len(errors) != 0 {
		return fmt.Errorf("%w: %v", ErrInvalidFields, strings.Join(errors, ". "))
	}

	return nil
}
