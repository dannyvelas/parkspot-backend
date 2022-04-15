package storage

import (
	"github.com/dannyvelas/lasvistas_api/models"
	"time"
)

type permit struct {
	PermitId   int    `db:"permit_id"`
	ResidentId string `db:"resident_id"`
	car
	StartTS     int64 `db:"start_ts"`
	EndTS       int64 `db:"end_ts"`
	RequestTS   int64 `db:"request_ts"`
	AffectsDays bool  `db:"affects_days"`
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
		StartDate:   time.Unix(permit.StartTS, 0),
		EndDate:     time.Unix(permit.EndTS, 0),
		RequestTS:   permit.RequestTS,
		AffectsDays: permit.AffectsDays,
	}
}
