package psql

import (
	"database/sql"
	"github.com/dannyvelas/parkspot-backend/models"
	"time"
)

type permit struct {
	PermitID        int            `db:"permit_id"`
	ResidentID      string         `db:"resident_id"`
	CarID           string         `db:"car_id"`
	LicensePlate    string         `db:"license_plate"`
	Color           string         `db:"color"`
	Make            sql.NullString `db:"make"`
	Model           sql.NullString `db:"model"`
	StartTS         int64          `db:"start_ts"`
	EndTS           int64          `db:"end_ts"`
	RequestTS       sql.NullInt64  `db:"request_ts"`
	AffectsDays     bool           `db:"affects_days"`
	ExceptionReason sql.NullString `db:"exception_reason"`
	CancelledTS     sql.NullInt64  `db:"cancelled_ts"`
}

func (permit permit) toModels() models.Permit {
	p := models.NewPermit(
		permit.PermitID,
		permit.ResidentID,
		permit.CarID,
		permit.LicensePlate,
		permit.Color,
		permit.Make.String,
		permit.Model.String,
		time.Unix(permit.StartTS, 0), // time.Unix() returns time in local tz
		time.Unix(permit.EndTS, 0),
		permit.RequestTS.Int64,
		permit.AffectsDays,
		permit.ExceptionReason.String,
	)

	if permit.CancelledTS.Valid {
		cancelledTS := permit.CancelledTS.Int64
		p.CancelledTS = &cancelledTS
	}

	return p
}

type permitSlice []permit

func (permits permitSlice) toModels() []models.Permit {
	modelsPermits := make([]models.Permit, 0, len(permits))
	for _, permit := range permits {
		modelsPermits = append(modelsPermits, permit.toModels())
	}
	return modelsPermits
}
