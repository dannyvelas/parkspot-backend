package models

import (
	"fmt"
	"regexp"
	"strings"
)

type CreateCar struct {
	LicensePlate string `json:"licensePlate"`
	Color        string `json:"color"`
	Make         string `json:"make"`
	Model        string `json:"model"`
}

func (createCar CreateCar) emptyFields() error {
	emptyFields := []string{}

	if createCar.LicensePlate == "" {
		emptyFields = append(emptyFields, "LicensePlate")
	}
	if createCar.Color == "" {
		emptyFields = append(emptyFields, "Color")
	}
	if createCar.Make == "" {
		emptyFields = append(emptyFields, "Make")
	}
	if createCar.Model == "" {
		emptyFields = append(emptyFields, "Model")
	}

	if len(emptyFields) > 0 {
		return fmt.Errorf("%w: %v", ErrEmptyFields, strings.Join(emptyFields, ", "))
	}

	return nil
}

func (createCar CreateCar) invalidFields() error {
	errors := []string{}

	if !regexp.MustCompile("^[A-Za-z0-9]+$").MatchString(createCar.LicensePlate) {
		errors = append(errors, "licensePlate can only be letters or numbers")
	}
	if len(createCar.LicensePlate) > 8 {
		errors = append(errors, "licensePlate can be maximum 8 characters")
	}

	if colorMakeModelErrors := invalidColorMakeModel(
		createCar.Color,
		createCar.Make,
		createCar.Model,
	); len(colorMakeModelErrors) != 0 {
		errors = append(errors, colorMakeModelErrors...)
	}

	if len(errors) > 0 {
		return fmt.Errorf("%w: %v", ErrInvalidFields, strings.Join(errors, ". "))
	}

	return nil
}

func (createCar CreateCar) Validate() error {
	if err := createCar.emptyFields(); err != nil {
		return err
	}

	if err := createCar.invalidFields(); err != nil {
		return err
	}

	return nil
}
