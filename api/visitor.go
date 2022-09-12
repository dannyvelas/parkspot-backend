package api

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/models"
	"strings"
	"time"
)

type newVisitorReq struct {
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	Relationship string    `json:"relationship"`
	IsForever    bool      `json:"isForever"`
	AccessStart  time.Time `json:"accessStart"`
	AccessEnd    time.Time `json:"accessEnd"`
}

func (newVisitorReq newVisitorReq) emptyFields() error {
	emptyFields := []string{}

	if newVisitorReq.FirstName == "" {
		emptyFields = append(emptyFields, "firstName")
	}
	if newVisitorReq.LastName == "" {
		emptyFields = append(emptyFields, "lastName")
	}
	if newVisitorReq.Relationship == "" {
		emptyFields = append(emptyFields, "relationship")
	}
	if !newVisitorReq.IsForever {
		if newVisitorReq.AccessStart.IsZero() {
			emptyFields = append(emptyFields, "accessStart")
		}
		if newVisitorReq.AccessEnd.IsZero() {
			emptyFields = append(emptyFields, "accessEnd")
		}
	}

	if len(emptyFields) > 0 {
		return fmt.Errorf("%w: %v", errEmptyFields, strings.Join(emptyFields, ", "))
	}

	return nil
}

func (newVisitorReq newVisitorReq) invalidFields() error {
	errors := []string{}

	if newVisitorReq.Relationship != "fam/fri" && newVisitorReq.Relationship != "contractor" {
		errors = append(errors, "relationship must be either \"fam/fri\" or \"contractor\"")
	}
	if newVisitorReq.Relationship == "contractor" && newVisitorReq.IsForever {
		errors = append(errors, "contractors cannot have an access date of `forever`")
	}
	if !newVisitorReq.IsForever {
		if newVisitorReq.AccessStart.After(newVisitorReq.AccessEnd) {
			errors = append(errors, "accessStart cannot be after accessEnd")
		}
		if newVisitorReq.AccessStart.Equal(newVisitorReq.AccessEnd) {
			errors = append(errors, "accessStart cannot be equal to accessEnd")
		}
		if newVisitorReq.AccessEnd.After(models.EndOfTime) {
			errors = append(errors, "accessEnd cannot be after 9999/12/31")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%w: %v", errInvalidFields, strings.Join(errors, ". "))
	}

	return nil
}

func (newVisitorReq newVisitorReq) validate() error {
	if err := newVisitorReq.emptyFields(); err != nil {
		return err
	}

	if err := newVisitorReq.invalidFields(); err != nil {
		return err
	}

	return nil
}
