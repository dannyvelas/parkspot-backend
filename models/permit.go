package models

import (
	"time"
)

type Permit struct {
	ID              int       `json:"id"`
	ResidentID      string    `json:"residentID"`
	CarID           string    `json:"carID"`
	LicensePlate    string    `json:"licensePlate"`
	Color           string    `json:"color"`
	Make            string    `json:"make"`
	Model           string    `json:"model"`
	StartDate       time.Time `json:"startDate"`
	EndDate         time.Time `json:"endDate"`
	RequestTS       int64     `json:"requestTS"` // int64: type used by time package for unix time
	AffectsDays     bool      `json:"affectsDays"`
	ExceptionReason string    `json:"exceptionReason,omitempty"`
}

func NewPermit(
	id int,
	residentID string,
	carID string,
	licensePlate string,
	color string,
	make string,
	model string,
	startDate time.Time,
	endDate time.Time,
	requestTS int64,
	affectsDays bool,
	exceptionReason string,
) Permit {
	return Permit{
		ID:              id,
		ResidentID:      residentID,
		CarID:           carID,
		LicensePlate:    licensePlate,
		Color:           color,
		Make:            make,
		Model:           model,
		StartDate:       startDate,
		EndDate:         endDate,
		RequestTS:       requestTS,
		AffectsDays:     affectsDays,
		ExceptionReason: exceptionReason,
	}
}

func (p Permit) Equal(other Permit) bool {
	if p.ID != other.ID {
		return false
	} else if p.ResidentID != other.ResidentID {
		return false
	} else if p.CarID != other.CarID {
		return false
	} else if p.LicensePlate != other.LicensePlate {
		return false
	} else if p.Color != other.Color {
		return false
	} else if p.Make != other.Make {
		return false
	} else if p.Model != other.Model {
		return false
	} else if !p.StartDate.Equal(other.StartDate) {
		return false
	} else if !p.EndDate.Equal(other.EndDate) {
		return false
	} else if p.RequestTS != other.RequestTS {
		return false
	} else if p.AffectsDays != other.AffectsDays {
		return false
	} else if p.ExceptionReason != other.ExceptionReason {
		return false
	}

	return true
}
