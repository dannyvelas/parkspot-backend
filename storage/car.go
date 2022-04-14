package storage

import (
	"database/sql"
)

type car struct {
	CarId        string         `db:"car_id"`
	LicensePlate string         `db:"license_plate"`
	Color        string         `db:"color"`
	Make         sql.NullString `db:"make"`
	Model        sql.NullString `db:"model"`
}
