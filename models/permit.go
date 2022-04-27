package models

import (
	"github.com/google/go-cmp/cmp"
	"time"
)

type Permit struct {
	Id          int64     `json:"id"`
	ResidentId  string    `json:"residentId"`
	Car         Car       `json:"car"`
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate"`
	RequestTS   int64     `json:"requestTS"`
	AffectsDays bool      `json:"affectsDays"`
}

func NewPermit(id int64, residentId string, car Car, startDate time.Time, endDate time.Time, requestTS int64, affectsDays bool) Permit {
	return Permit{
		Id:          id,
		ResidentId:  residentId,
		Car:         car,
		StartDate:   startDate,
		EndDate:     endDate,
		RequestTS:   requestTS,
		AffectsDays: affectsDays,
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
	}

	return true
}

type CreatePermit struct {
	ResidentId  string    `json:"residentId"`
	CreateCar   CreateCar `json:"car"`
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate"`
	RequestTS   int64     `json:"requestTS"`
	AffectsDays bool      `json:"affectsDays"`
}

func NewCreatePermit(residentId string, createCar CreateCar, startDate time.Time, endDate time.Time, requestTS int64, affectsDays bool) CreatePermit {
	return CreatePermit{
		ResidentId:  residentId,
		CreateCar:   createCar,
		StartDate:   startDate,
		EndDate:     endDate,
		RequestTS:   requestTS,
		AffectsDays: affectsDays,
	}
}

func (createPermit CreatePermit) ToPermit(permitId int64, carId string) Permit {
	return NewPermit(
		permitId,
		createPermit.ResidentId,
		createPermit.CreateCar.ToCar(carId),
		createPermit.StartDate,
		createPermit.EndDate,
		createPermit.RequestTS,
		createPermit.AffectsDays)
}
