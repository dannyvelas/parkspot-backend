package models

import (
	"fmt"
	"regexp"
)

type Car struct {
	Id           string `json:"id"`
	LicensePlate string `json:"licensePlate"`
	Color        string `json:"color"`
	Make         string `json:"make"`
	Model        string `json:"model"`
}

func NewCar(id string, licensePlate string, color string, make string, model string) Car {
	return Car{
		Id:           id,
		LicensePlate: licensePlate,
		Color:        color,
		Make:         make,
		Model:        model,
	}
}

func (self Car) Equal(other Car) bool {
	if self.Id != other.Id {
		return false
	} else if self.LicensePlate != other.LicensePlate {
		return false
	} else if self.Color != other.Color {
		return false
	} else if self.Make != other.Make {
		return false
	} else if self.Model != other.Model {
		return false
	}

	return true
}

func (car Car) Validate() error {
	if !regexp.MustCompile("^[A-Za-z0-9]+$").MatchString(car.LicensePlate) {
		return fmt.Errorf("Invalid car: licensePlate can only be letters or numbers")
	}

	if len(car.LicensePlate) > 8 {
		return fmt.Errorf("Invalid car: licensePlate can be maximum 8 characters")
	}

	if !regexp.MustCompile("^[A-Za-z]+$").MatchString(car.Color) {
		return fmt.Errorf("Invalid car: color must be one word only letters")
	}

	makeAndModelRe := regexp.MustCompile("^[A-Za-z0-9 -]+$")
	if !makeAndModelRe.MatchString(car.Make) {
		return fmt.Errorf("Invalid car: make can only have spaces, letters, numbers, and dashes")
	}

	if !makeAndModelRe.MatchString(car.Model) {
		return fmt.Errorf("Invalid car: model can only have spaces, letters, numbers, and dashes")
	}

	return nil
}
