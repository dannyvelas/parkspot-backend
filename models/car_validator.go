package models

import (
	"github.com/dannyvelas/lasvistas_api/errs"
	"regexp"
	"strings"
)

type carFieldValidator struct {
	licensePlateRe    *regexp.Regexp
	colorRe           *regexp.Regexp
	makeModelRe       *regexp.Regexp
	validateAmtDaysFn func(*int) error
}

var (
	CreateCarValidator = carFieldValidator{
		regexp.MustCompile("^[A-Za-z0-9]+$"),
		regexp.MustCompile("^[A-Za-z]+$"),
		regexp.MustCompile("^[A-Za-z0-9 -]+$"),
		nil,
	}
	EditCarValidator = carFieldValidator{
		regexp.MustCompile("^[A-Za-z0-9]*$"),
		regexp.MustCompile("^[A-Za-z]*$"),
		regexp.MustCompile("^[A-Za-z0-9 -]*$"),
		validateEditAmtDays,
	}
)

func (v carFieldValidator) Run(car Car) *errs.ApiErr {
	var errors []string

	if !v.licensePlateRe.MatchString(car.LicensePlate) {
		errors = append(errors, "licensePlate can only be letters or numbers")
	}
	if len(car.LicensePlate) > 8 {
		errors = append(errors, "licensePlate can be maximum 8 characters")
	}
	if !v.colorRe.MatchString(car.Color) {
		errors = append(errors, "color must be one word only letters")
	}
	if !v.makeModelRe.MatchString(car.Make) {
		errors = append(errors, "make can only have spaces, letters, numbers, and dashes")
	}
	if !v.makeModelRe.MatchString(car.Model) {
		errors = append(errors, "model can only have spaces, letters, numbers, and dashes")
	}
	if v.validateAmtDaysFn != nil {
		if err := v.validateAmtDaysFn(car.AmtParkingDaysUsed); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) != 0 {
		return errs.InvalidFields(strings.Join(errors, ". "))
	}

	return nil
}

func validateEditAmtDays(amtDays *int) error {
	if amtDays == nil {
		return nil
	}

	if *amtDays < 0 {
		return errs.InvalidFields("amtDays cannot be lower than 0")
	}

	return nil
}
