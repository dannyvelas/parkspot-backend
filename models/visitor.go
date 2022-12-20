package models

import (
	"fmt"
	"strings"
	"time"
)

type Visitor struct {
	ID           string    `json:"id"`
	ResidentID   string    `json:"residentID"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	Relationship string    `json:"relationship"`
	AccessStart  time.Time `json:"accessStart"`
	AccessEnd    time.Time `json:"accessEnd"`
}

func NewVisitor(
	id string,
	residentID string,
	firstName string,
	lastName string,
	relationship string,
	accessStart time.Time,
	accessEnd time.Time,
) Visitor {
	return Visitor{
		ID:           id,
		ResidentID:   residentID,
		FirstName:    firstName,
		LastName:     lastName,
		Relationship: relationship,
		AccessStart:  accessStart,
		AccessEnd:    accessEnd,
	}
}

func (m Visitor) ValidateCreation() error {
	if err := m.emptyFields(); err != nil {
		return err
	}

	if err := m.invalidFields(); err != nil {
		return err
	}

	return nil
}

func (m Visitor) emptyFields() error {
	emptyFields := []string{}

	if m.FirstName == "" {
		emptyFields = append(emptyFields, "firstName")
	}
	if m.LastName == "" {
		emptyFields = append(emptyFields, "lastName")
	}
	if m.Relationship == "" {
		emptyFields = append(emptyFields, "relationship")
	}
	if m.AccessStart.IsZero() {
		emptyFields = append(emptyFields, "accessStart")
	}
	if m.AccessEnd.IsZero() {
		emptyFields = append(emptyFields, "accessEnd")
	}

	if len(emptyFields) > 0 {
		return fmt.Errorf("%w: %v", ErrEmptyFields, strings.Join(emptyFields, ", "))
	}

	return nil
}

func (m Visitor) invalidFields() error {
	errors := []string{}

	if m.Relationship != "fam/fri" && m.Relationship != "contractor" {
		errors = append(errors, "relationship must be either \"fam/fri\" or \"contractor\"")
	}
	if m.AccessStart.After(m.AccessEnd) {
		errors = append(errors, "accessStart cannot be after accessEnd")
	}
	if m.AccessStart.Equal(m.AccessEnd) {
		errors = append(errors, "accessStart cannot be equal to accessEnd")
	}
	if m.AccessEnd.After(EndOfTime) {
		errors = append(errors, "accessEnd cannot be after 9999/12/31")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%w: %v", ErrInvalidFields, strings.Join(errors, ". "))
	}

	return nil
}
