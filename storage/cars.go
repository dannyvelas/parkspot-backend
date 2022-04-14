package storage

import (
	"database/sql"
)

type Car struct {
	CarId        string `db:"car_id"`
	LicensePlate string `db:"license_plate"`
	Color        string
	Make         sql.NullString
	Model        sql.NullString
}
