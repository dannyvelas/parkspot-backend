package storage

import (
	"github.com/dannyvelas/lasvistas_api/models"
)

type resident struct {
	ID                 string `db:"id"`
	FirstName          string `db:"first_name"`
	LastName           string `db:"last_name"`
	Phone              string `db:"phone"`
	Email              string `db:"email"`
	Password           string `db:"password"`
	UnlimDays          bool   `db:"unlim_days"`
	AmtParkingDaysUsed int    `db:"amt_parking_days_used"`
	TokenVersion       int    `db:"token_version"`
}

func (resident resident) toModels() models.Resident {
	return models.NewResident(
		resident.ID,
		resident.FirstName,
		resident.LastName,
		resident.Phone,
		resident.Email,
		"",
		&resident.UnlimDays,
		&resident.AmtParkingDaysUsed,
		resident.TokenVersion,
	)
}

type residentSlice []resident

func (residents residentSlice) toModels() []models.Resident {
	modelsResidents := make([]models.Resident, 0, len(residents))
	for _, resident := range residents {
		modelsResidents = append(modelsResidents, resident.toModels())
	}
	return modelsResidents
}
