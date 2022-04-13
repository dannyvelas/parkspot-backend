package storage

import (
	"fmt"
)

type PermitsRepo struct {
	database Database
}

func NewPermitsRepo(database Database) PermitsRepo {
	return PermitsRepo{database: database}
}

func (permitsRepo PermitsRepo) GetActive(limit, offset uint) ([]Permit, error) {
	const query = `
    SELECT
      permits.id AS id,
      permits.resident_id AS resident_id,
      cars.id,
      cars.license_plate,
      cars.color,
      cars.make,
      cars.model,
      permits.start_date,
      permits.end_date,
      permits.request_ts,
      permits.affects_days
    FROM permits
    LEFT JOIN cars ON
      permits.car_id = cars.id 
    WHERE
      permits.start_date <= NOW()
      AND permits.end_date >= NOW()
    LIMIT $1
    OFFSET $2
  `

	permits := []Permit{}
	err := permitsRepo.database.driver.Select(&permits, query, getBoundedLimit(limit), offset)
	if err != nil {
		return nil, fmt.Errorf("permits_repo: GetActive: %v", newError(ErrDatabaseQuery, err))
	}

	return permits, nil
}

func (permitsRepo PermitsRepo) GetAll(limit, offset uint) ([]Permit, error) {
	const query = `
    SELECT
      permits.id,
      permits.resident_id,
      cars.id AS cars_id,
      cars.license_plate AS license_plate,
      cars.color,
      cars.make,
      cars.model,
      permits.start_date,
      permits.end_date,
      permits.request_ts,
      permits.affects_days
    FROM permits
    LEFT JOIN cars ON
      permits.car_id = cars.id
    LIMIT $1
    OFFSET $2
  `

	permits := []Permit{}
	err := permitsRepo.database.driver.Select(&permits, query, getBoundedLimit(limit), offset)
	if err != nil {
		return nil, fmt.Errorf("permits_repo: GetAll: %v", newError(ErrDatabaseQuery, err))
	}

	return permits, nil
}

func (permitsRepo PermitsRepo) deleteAll() (int64, error) {
	query := "DELETE FROM permits"
	res, err := permitsRepo.database.driver.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("permits_repo: deleteAll: %v", newError(ErrDatabaseExec, err))
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("permits_repo: deleteAll: %v", newError(ErrGetRowsAffected, err))
	}

	return rowsAffected, nil
}
