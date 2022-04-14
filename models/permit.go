package models

import (
	"time"
)

type Permit struct {
	Id          int       `json:"id"`
	ResidentId  string    `json:"resident_id"`
	Car         Car       `json:"car"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	RequestTS   int       `json:"request_ts"`
	AffectsDays bool      `json:"affects_days"`
}
