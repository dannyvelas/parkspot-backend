package validator

import (
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"regexp"
	"strings"
)

type residentValidator struct {
	firstLastRe       *regexp.Regexp
	phoneRe           *regexp.Regexp
	emailRe           *regexp.Regexp
	passwordEmptyOk   bool
	validateAmtDaysFn func(*int) error
}

var (
	CreateResident = residentValidator{
		regexp.MustCompile("^[A-Za-z ]+$"),
		regexp.MustCompile("^\\d{1,20}$"),
		regexp.MustCompile("^.+@.+$"),
		false,
		nil,
	}
	EditResident = residentValidator{
		regexp.MustCompile("^[A-Za-z ]*$"),
		regexp.MustCompile("^\\d{0,20}$"),
		regexp.MustCompile("^(.+@.+$|)$"),
		true,
		validateEditAmtDays,
	}
)

func (v residentValidator) Run(resident models.Resident) *errs.ApiErr {
	var errors []string

	if err := models.IsResidentID(resident.ID); err != nil {
		errors = append(errors, err.Error())
	}
	if !v.firstLastRe.MatchString(resident.FirstName) {
		errors = append(errors, "first name can only be alphabetic letters and spaces")
	}
	if !v.firstLastRe.MatchString(resident.LastName) {
		errors = append(errors, "last name can only be alphabetic letters and spaces")
	}
	if !v.phoneRe.MatchString(resident.Phone) {
		errors = append(errors, "phone must be only numbers, at most 20")
	}
	if !v.emailRe.MatchString(resident.Email) {
		errors = append(errors, "email must be a sequence of characters separated by an '@' character.")
	}
	if !v.passwordEmptyOk && resident.Password == "" {
		errors = append(errors, "password must not be empty")
	}
	if v.validateAmtDaysFn != nil {
		if err := v.validateAmtDaysFn(resident.AmtParkingDaysUsed); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) != 0 {
		return errs.InvalidFields(strings.Join(errors, ". "))
	}

	return nil
}
