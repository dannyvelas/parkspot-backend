package validator

import (
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/util"
	"regexp"
	"strings"
)

type carValidator struct {
	validateIDFn      func(string) error
	validateResIDFn   func(string) error
	licensePlateRe    *regexp.Regexp
	colorRe           *regexp.Regexp
	makeModelRe       *regexp.Regexp
	validateAmtDaysFn func(*int) error
}

var (
	CreateCar = carValidator{
		validateCarID,
		models.IsResidentID,
		regexp.MustCompile("^[A-Za-z0-9]+$"),
		regexp.MustCompile("^[A-Za-z]+$"),
		regexp.MustCompile("^[A-Za-z0-9 -]+$"),
		nil,
	}
	EditCar = carValidator{
		nil,
		nil,
		regexp.MustCompile("^[A-Za-z0-9]*$"),
		regexp.MustCompile("^[A-Za-z]*$"),
		regexp.MustCompile("^[A-Za-z0-9 -]*$"),
		validateEditAmtDays,
	}
)

func (v carValidator) Run(car models.Car) *errs.ApiErr {
	var errors []string

	if v.validateIDFn != nil {
		if err := v.validateIDFn(car.ID); err != nil {
			errors = append(errors, err.Error())
		}
	}
	if v.validateResIDFn != nil {
		if err := v.validateResIDFn(car.ResidentID); err != nil {
			errors = append(errors, err.Error())
		}
	}
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

func validateCarID(carID string) error {
	// ID field is optional, but if provided, ensure it is UUID
	if carID != "" && !util.IsUUIDV4(carID) {
		return errs.IDNotUUID
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
