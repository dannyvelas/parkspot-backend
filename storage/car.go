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
	return models.Car{
		Id:           car.CarId,
		LicensePlate: car.LicensePlate,
		Color:        car.Color,
		Make:         car.Make.String,
		Model:        car.Model.String,
	}
}
