package storage

import (
	"database/sql"
	"github.com/dannyvelas/lasvistas_api/models"
	"time"
)

type permit struct {
	PermitId   int64  `db:"permit_id"`
	ResidentId string `db:"resident_id"`
	car
	StartTS         int64          `db:"start_ts"`
	EndTS           int64          `db:"end_ts"`
	RequestTS       sql.NullInt64  `db:"request_ts"`
	AffectsDays     bool           `db:"affects_days"`
	ExceptionReason sql.NullString `db:"exception_reason"`
}

func (permit permit) toModels() models.Permit {
	return models.NewPermit(
		permit.PermitId,
		permit.ResidentId,
		permit.car.toModels(),
		time.Unix(permit.StartTS, 0),
		time.Unix(permit.EndTS, 0),
		permit.RequestTS.Int64,
		permit.AffectsDays,
		permit.ExceptionReason.String,
	)
}

type permitSlice []permit

func (permits permitSlice) toModels() []models.Permit {
	modelsPermits := make([]models.Permit, 0, len(permits))
	for _, permit := range permits {
		modelsPermits = append(modelsPermits, permit.toModels())
	}
	return modelsPermits
}
