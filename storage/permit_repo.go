package storage

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/models"
)

const permitVersion = 4

type PermitRepo struct {
	database Database
}

func NewPermitRepo(database Database) PermitRepo {
	return PermitRepo{database: database}
}

func (permitRepo PermitRepo) GetActive(limit, offset uint) ([]models.Permit, error) {
	const query = `
    SELECT
      permit.id AS permit_id,
      permit.resident_id,
      car.id AS car_id,
      car.license_plate,
      car.color,
      car.make,
      car.model,
      permit.start_ts,
      permit.end_ts,
      permit.request_ts,
      permit.affects_days
    FROM permit
    LEFT JOIN car ON
      permit.car_id = car.id 
    WHERE
      permit.start_ts <= extract(epoch from now())
      AND permit.end_ts >= extract(epoch from now())
    LIMIT $1
    OFFSET $2
  `

	permits := permitSlice{}
	err := permitRepo.database.driver.Select(&permits, query, getBoundedLimit(limit), offset)
	if err != nil {
		return nil, fmt.Errorf("permit_repo: GetActive: %v", newError(ErrDatabaseQuery, err))
	}

	return permits.toModels(), nil
}

func (permitRepo PermitRepo) GetAll(limit, offset uint) ([]models.Permit, error) {
	const query = `
    SELECT
      permit.id AS permit_id,
      permit.resident_id,
      car.id AS car_id,
      car.license_plate,
      car.color,
      car.make,
      car.model,
      permit.start_ts,
      permit.end_ts,
      permit.request_ts,
      permit.affects_days
    FROM permit
    LEFT JOIN car ON
      permit.car_id = car.id
    LIMIT $1
    OFFSET $2
  `

	permits := permitSlice{}
	err := permitRepo.database.driver.Select(&permits, query, getBoundedLimit(limit), offset)
	if err != nil {
		return nil, fmt.Errorf("permit_repo: GetAll: %v", newError(ErrDatabaseQuery, err))
	}

	return permits.toModels(), nil
}

func (permitRepo PermitRepo) deleteAll() (int64, error) {
	query := "DELETE FROM permit"
	res, err := permitRepo.database.driver.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("permit_repo: deleteAll: %v", newError(ErrDatabaseExec, err))
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("permit_repo: deleteAll: %v", newError(ErrGetRowsAffected, err))
	}

	return rowsAffected, nil
}
