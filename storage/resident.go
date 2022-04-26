package storage

import (
	"github.com/dannyvelas/lasvistas_api/models"
)

type resident struct {
	Id                 string `db:"id"`
	FirstName          string `db:"first_name"`
	LastName           string `db:"last_name"`
	Phone              string `db:"phone"`
	Email              string `db:"email"`
	Password           string `db:"password"`
	UnlimDays          bool   `db:"unlim_days"`
	AmtParkingDaysUsed int    `db:"amt_parking_days_used"`
}

func (resident resident) toModels() models.Resident {
	return models.Resident{
		Id:                 resident.Id,
		FirstName:          resident.FirstName,
		LastName:           resident.LastName,
		Phone:              resident.Phone,
		Email:              resident.Email,
		Password:           resident.Password,
		UnlimDays:          resident.UnlimDays,
		AmtParkingDaysUsed: resident.AmtParkingDaysUsed,
	}
}
