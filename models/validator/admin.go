package validator

import (
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"regexp"
	"strings"
)

type adminValidator struct {
	firstLastRe *regexp.Regexp
	emailRe     *regexp.Regexp
}

var (
	CreateAdmin = adminValidator{
		regexp.MustCompile("^[A-Za-z ]+$"),
		regexp.MustCompile("^.+@.+$"),
	}
)

func (v adminValidator) Run(admin models.Admin) *errs.ApiErr {
	var errors []string

	if admin.ID == "" {
		errors = append(errors, "id cannot be empty")
	}
	if !v.firstLastRe.MatchString(admin.FirstName) {
		errors = append(errors, "first name can only be alphabetic letters and spaces")
	}
	if !v.firstLastRe.MatchString(admin.LastName) {
		errors = append(errors, "last name can only be alphabetic letters and spaces")
	}
	if !v.emailRe.MatchString(admin.Email) {
		errors = append(errors, "email must be a sequence of characters separated by an '@' character.")
	}

	if len(errors) != 0 {
		return errs.InvalidFields(strings.Join(errors, ". "))
	}

	return nil
}
