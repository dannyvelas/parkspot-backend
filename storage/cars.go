package storage

type Car struct {
	CarId        string `db:"cars_id"`
	LicensePlate string `db:"license_plate"`
	Color        string
	Make         string
	Model        string
}
