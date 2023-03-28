package models

import (
	"regexp"
)

func getLPColorMakeModelErrors(licensePlate, color, make, model string) []string {
	errors := []string{}

	if !regexp.MustCompile("^[A-Za-z0-9]+$").MatchString(licensePlate) {
		errors = append(errors, "licensePlate can only be letters or numbers")
	}
	if len(licensePlate) > 8 {
		errors = append(errors, "licensePlate can be maximum 8 characters")
	}

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
