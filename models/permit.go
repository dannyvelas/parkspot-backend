package models

import (
	"github.com/google/go-cmp/cmp"
	"time"
)

type Permit struct {
	Id          int       `json:"id"`
	ResidentId  string    `json:"resident_id"`
	Car         Car       `json:"car"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	RequestTS   int64     `json:"request_ts"`
	AffectsDays bool      `json:"affects_days"`
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
