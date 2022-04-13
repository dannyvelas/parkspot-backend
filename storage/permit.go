package storage

import (
	"time"
)

type Permit struct {
	Id         int    `db:"id"`
	ResidentId string `db:"resident_id"`
	Car
	StartDate   time.Time `db:"start_date"`
	EndDate     time.Time `db:"end_date"`
	RequestTS   int       `db:"request_ts"`
	AffectsDays bool      `db:"affects_days"`
}
