package models

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"time"
)

type Permit struct {
	Id          int       `json:"id"`
	ResidentId  string    `json:"residentId"`
	Car         Car       `json:"car"`
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate"`
	RequestTS   int64     `json:"requestTS"`
	AffectsDays bool      `json:"affectsDays"`
}

func NewPermit(id int, residentId string, car Car, startString string, endString string, requestTS int64, affectsDays bool) (Permit, error) {
	const dateFormat = "2006-01-02"

	startDate, err := time.ParseInLocation(dateFormat, startString, time.Local)
	if err != nil {
		return Permit{}, fmt.Errorf("Error parsing startDate: %v", err)
	}

	endDate, err := time.ParseInLocation(dateFormat, endString, time.Local)
	if err != nil {
		return Permit{}, fmt.Errorf("Error parsing endDate: %v", err)
	}

	return Permit{
		Id:          id,
		ResidentId:  residentId,
		Car:         car,
		StartDate:   startDate,
		EndDate:     endDate,
		RequestTS:   requestTS,
		AffectsDays: affectsDays,
	}, nil
}

func (self Permit) Equal(other Permit) bool {
	if self.Id != other.Id {
		return false
	} else if self.ResidentId != other.ResidentId {
		return false
	} else if !cmp.Equal(self.Car, other.Car) {
		return false
	} else if !self.StartDate.Equal(other.StartDate) {
		return false
	} else if !self.EndDate.Equal(other.EndDate) {
		return false
	} else if self.RequestTS != other.RequestTS {
		return false
	} else if self.AffectsDays != other.AffectsDays {
		return false
	}

	return true
}
