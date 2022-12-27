package storage

import (
	"database/sql"
	"github.com/dannyvelas/lasvistas_api/models"
)

type car struct {
	CarID              string         `db:"car_id"`
	ResidentID         string         `db:"resident_id"`
	LicensePlate       string         `db:"license_plate"`
	Color              string         `db:"color"`
	Make               sql.NullString `db:"make"`
	Model              sql.NullString `db:"model"`
	AmtParkingDaysUsed int            `db:"amt_parking_days_used"`
}

func (car car) toModels() models.Car {
	return models.NewCar(
		car.CarID,
		car.ResidentID,
		car.LicensePlate,
		car.Color,
		car.Make.String,
		car.Model.String,
		car.AmtParkingDaysUsed)
}
