package models

import (
	"github.com/google/go-cmp/cmp"
	"time"
)

type Permit struct {
	Id              int       `json:"id"`
	ResidentId      string    `json:"residentId"`
	Car             Car       `json:"car"`
	StartDate       time.Time `json:"startDate"`
	EndDate         time.Time `json:"endDate"`
	RequestTS       int64     `json:"requestTS"` // int64: type used by time package for unix time
	AffectsDays     bool      `json:"affectsDays"`
	ExceptionReason string    `json:"exceptionReason"`
}

func NewPermit(id int, residentId string, car Car, startDate time.Time, endDate time.Time, requestTS int64, affectsDays bool, exceptionReason string) Permit {
	return Permit{
		Id:              id,
		ResidentId:      residentId,
		Car:             car,
		StartDate:       startDate,
		EndDate:         endDate,
		RequestTS:       requestTS,
		AffectsDays:     affectsDays,
		ExceptionReason: exceptionReason,
	}
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
	} else if self.ExceptionReason != other.ExceptionReason {
		return false
	}

	return true
}

type NewPermitArgs struct {
	ResidentId      string
	CarId           string
	StartDate       time.Time
	EndDate         time.Time
	AffectsDays     bool
	ExceptionReason string
}

func NewNewPermitArgs(residentId string, carId string, startDate time.Time, endDate time.Time, affectsDays bool, exceptionReason string) NewPermitArgs {
	return NewPermitArgs{
		ResidentId:      residentId,
		CarId:           carId,
		StartDate:       startDate,
		EndDate:         endDate,
		AffectsDays:     affectsDays,
		ExceptionReason: exceptionReason,
	}
}
