package storage

import (
	"github.com/dannyvelas/lasvistas_api/models"
	"time"
)

type permit struct {
	PermitId   int    `db:"id"`
	ResidentId string `db:"resident_id"`
	car
	StartDate   time.Time `db:"start_date"`
	EndDate     time.Time `db:"end_date"`
	RequestTS   int       `db:"request_ts"`
	AffectsDays bool      `db:"affects_days"`
}

func (permit permit) toModels() models.Permit {
	return models.Permit{
		Id:         permit.PermitId,
		ResidentId: permit.ResidentId,
		Car: models.Car{
			Id:           permit.CarId,
			LicensePlate: permit.LicensePlate,
			Color:        permit.Color,
			Make:         permit.Make.String,
			Model:        permit.Model.String,
		},
		StartDate:   permit.StartDate,
		EndDate:     permit.EndDate,
		RequestTS:   permit.RequestTS,
		AffectsDays: permit.AffectsDays,
	}
}
