package models

import (
	"time"
)

var EndOfTime = time.Date(9999, time.December, 31, 0, 0, 0, 0, time.Local)

type Visitor struct {
	Id           string    `json:"id"`
	ResidentId   string    `json:"residentId"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	Relationship string    `json:"relationship"`
	AccessStart  time.Time `json:"accessStart"`
	AccessEnd    time.Time `json:"accessEnd"`
}

func NewVisitor(
	id string,
	residentId string,
	firstName string,
	lastName string,
	relationship string,
	accessStart time.Time,
	accessEnd time.Time,
) Visitor {
	return Visitor{
		Id:           id,
		ResidentId:   residentId,
		FirstName:    firstName,
		LastName:     lastName,
		Relationship: relationship,
		AccessStart:  accessStart,
		AccessEnd:    accessEnd,
	}
}
