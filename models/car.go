package models

import (
	"fmt"
	"regexp"
	"strings"
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
	errors := []string{}
	if !regexp.MustCompile("^[A-Za-z0-9]+$").MatchString(car.LicensePlate) {
		errors = append(errors, "licensePlate can only be letters or numbers")
	}

	if len(car.LicensePlate) > 8 {
		errors = append(errors, "licensePlate can be maximum 8 characters")
	}

	if !regexp.MustCompile("^[A-Za-z]+$").MatchString(car.Color) {
		errors = append(errors, "color must be one word only letters")
	}

	makeAndModelRe := regexp.MustCompile("^[A-Za-z0-9 -]+$")
	if !makeAndModelRe.MatchString(car.Make) {
		errors = append(errors, "make can only have spaces, letters, numbers, and dashes")
	}

	if !makeAndModelRe.MatchString(car.Model) {
		errors = append(errors, "model can only have spaces, letters, numbers, and dashes")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%v", strings.Join(errors, ". "))
	}

	return nil
}

func (car Car) hasEmptyValue() bool {
	if car.Id == "" {
		return true
	} else if car.LicensePlate == "" {
		return true
	} else if car.Color == "" {
		return true
	} else if car.Make == "" {
		return true
	} else if car.Model == "" {
		return true
	}

	return false
}
