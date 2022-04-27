package storage

import (
	"database/sql"
	"github.com/dannyvelas/lasvistas_api/models"
)

type car struct {
	CarId        string         `db:"car_id"`
	LicensePlate string         `db:"license_plate"`
	Color        string         `db:"color"`
	Make         sql.NullString `db:"make"`
	Model        sql.NullString `db:"model"`
}

func (car car) toModels() models.Car {
	return models.NewCar(
		car.CarId,
		car.LicensePlate,
		car.Color,
		car.Make.String,
		car.Model.String,
	)
}
