package models

import (
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
