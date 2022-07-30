package api

import (
	"fmt"
	"regexp"
	"strings"
)

// new car request

type newCarReq struct {
	LicensePlate string `json:"licensePlate"`
	Color        string `json:"color"`
	Make         string `json:"make"`
	Model        string `json:"model"`
}

func (newCarReq newCarReq) emptyFields() error {
	emptyFields := []string{}

	if newCarReq.LicensePlate == "" {
		emptyFields = append(emptyFields, "LicensePlate")
	}
	if newCarReq.Color == "" {
		emptyFields = append(emptyFields, "Color")
	}
	if newCarReq.Make == "" {
		emptyFields = append(emptyFields, "Make")
	}
	if newCarReq.Model == "" {
		emptyFields = append(emptyFields, "Model")
	}

	if len(emptyFields) > 0 {
		return fmt.Errorf("%w: %v", errEmptyFields, strings.Join(emptyFields, ", "))
	}

	return nil
}

func (newCarReq newCarReq) invalidFields() error {
	errors := []string{}

	if !regexp.MustCompile("^[A-Za-z0-9]+$").MatchString(newCarReq.LicensePlate) {
		errors = append(errors, "licensePlate can only be letters or numbers")
	}
	if len(newCarReq.LicensePlate) > 8 {
		errors = append(errors, "licensePlate can be maximum 8 characters")
	}

	if colorMakeModelErrors := invalidColorMakeModel(
		newCarReq.Color,
		newCarReq.Make,
		newCarReq.Model,
	); len(colorMakeModelErrors) != 0 {
		errors = append(errors, colorMakeModelErrors...)
	}

	if len(errors) > 0 {
		return fmt.Errorf("%w: %v", errInvalidFields, strings.Join(errors, ". "))
	}

	return nil
}

func (newCarReq newCarReq) validate() error {
	if err := newCarReq.emptyFields(); err != nil {
		return err
	}

	if err := newCarReq.invalidFields(); err != nil {
		return err
	}

	return nil
}

// edit car request

type editCarReq struct {
	Color string `json:"color"`
	Make  string `json:"make"`
	Model string `json:"model"`
}

func (editCarReq editCarReq) validate() error {
	if editCarReq.Color == "" && editCarReq.Make == "" && editCarReq.Model == "" {
		return fmt.Errorf("%w: %v", errEmptyFields, "all edit fields cannot be empty")
	}

	if colorMakeModelErrors := invalidColorMakeModel(
		editCarReq.Color,
		editCarReq.Make,
		editCarReq.Model,
	); len(colorMakeModelErrors) != 0 {
		errors := colorMakeModelErrors
		return fmt.Errorf("%w: %v", errInvalidFields, strings.Join(errors, ". "))
	}

	return nil
}

// helpers

func invalidColorMakeModel(color, make, model string) []string {
	errors := []string{}

	if !regexp.MustCompile("^[A-Za-z]*$").MatchString(color) {
		errors = append(errors, "color must be one word only letters")
	}

	makeAndModelRe := regexp.MustCompile("^[A-Za-z0-9 -]*$")
	if !makeAndModelRe.MatchString(make) {
		errors = append(errors, "make can only have spaces, letters, numbers, and dashes")
	}

	if !makeAndModelRe.MatchString(model) {
		errors = append(errors, "model can only have spaces, letters, numbers, and dashes")
	}

	return errors
}
