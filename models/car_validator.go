package models

import (
	"github.com/dannyvelas/lasvistas_api/errs"
	"regexp"
	"strings"
)

type CarFieldValidator struct {
	licensePlateRe *regexp.Regexp
	colorRe        *regexp.Regexp
	makeModelRe    *regexp.Regexp
}

func NewCarFieldValidator(emptyOk bool) CarFieldValidator {
	if emptyOk {
		return CarFieldValidator{
			regexp.MustCompile("^[A-Za-z0-9]*$"),
			regexp.MustCompile("^[A-Za-z]*$"),
			regexp.MustCompile("^[A-Za-z0-9 -]*$"),
		}
	}
	return CarFieldValidator{
		regexp.MustCompile("^[A-Za-z0-9]+$"),
		regexp.MustCompile("^[A-Za-z]+$"),
		regexp.MustCompile("^[A-Za-z0-9 -]+$"),
	}
}

func (v CarFieldValidator) Validate(licensePlate, color, make, model string) *errs.ApiErr {
	errors := []string{}

	if !v.licensePlateRe.MatchString(licensePlate) {
		errors = append(errors, "licensePlate can only be letters or numbers")
	}
	if len(licensePlate) > 8 {
		errors = append(errors, "licensePlate can be maximum 8 characters")
	}
	if !v.colorRe.MatchString(color) {
		errors = append(errors, "color must be one word only letters")
	}
	if !v.makeModelRe.MatchString(make) {
		errors = append(errors, "make can only have spaces, letters, numbers, and dashes")
	}
	if !v.makeModelRe.MatchString(model) {
		errors = append(errors, "model can only have spaces, letters, numbers, and dashes")
	}

	if len(errors) != 0 {
		return errs.InvalidFields(strings.Join(errors, ". "))
	}

	return nil
}
