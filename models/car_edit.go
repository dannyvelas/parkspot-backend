package models

import (
	"fmt"
	"regexp"
	"strings"
)

type EditCar struct {
	Color string `json:"color"`
	Make  string `json:"make"`
	Model string `json:"model"`
}

func (editCar EditCar) Validate() error {
	if editCar.Color == "" && editCar.Make == "" && editCar.Model == "" {
		return fmt.Errorf("%w: %v", ErrEmptyFields, "all edit fields cannot be empty")
	}

	if colorMakeModelErrors := invalidColorMakeModel(
		editCar.Color,
		editCar.Make,
		editCar.Model,
	); len(colorMakeModelErrors) != 0 {
		errors := colorMakeModelErrors
		return fmt.Errorf("%w: %v", ErrInvalidFields, strings.Join(errors, ". "))
	}

	return nil
}

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
