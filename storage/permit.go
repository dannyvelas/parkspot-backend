package storage

import (
	"database/sql"
	"github.com/dannyvelas/lasvistas_api/models"
	"time"
)

type permit struct {
	PermitId   int    `db:"permit_id"`
	ResidentId string `db:"resident_id"`
	car
	StartTS     int64         `db:"start_ts"`
	EndTS       int64         `db:"end_ts"`
	RequestTS   sql.NullInt64 `db:"request_ts"`
	AffectsDays bool          `db:"affects_days"`
}

func (permit permit) toModels() models.Permit {
	return models.Permit{
		Id: permit.PermitId,
		PermitFields: models.PermitFields{
			ResidentId:  permit.ResidentId,
			Car:         permit.car.toModels(),
			StartDate:   time.Unix(permit.StartTS, 0),
			EndDate:     time.Unix(permit.EndTS, 0),
			RequestTS:   permit.RequestTS.Int64,
			AffectsDays: permit.AffectsDays,
		},
	}
}

type permitSlice []permit

func (permits permitSlice) toModels() []models.Permit {
	modelsPermits := make([]models.Permit, 0, len(permits))
	for _, permit := range permits {
		modelsPermits = append(modelsPermits, permit.toModels())
	}
	return modelsPermits
}
