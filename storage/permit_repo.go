package storage

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/slice"
)

type PermitRepo struct {
	database Database
}

func NewPermitRepo(database Database) PermitRepo {
	return PermitRepo{database: database}
}

func (permitRepo PermitRepo) GetActive(limit, offset uint) ([]models.Permit, error) {
	const query = `
    SELECT
      permit.id,
      permit.resident_id,
      car.id AS car_id,
      car.license_plate,
      car.color,
      car.make,
      car.model,
      permit.start_date,
      permit.end_date,
      permit.request_ts,
      permit.affects_days
    FROM permit
    LEFT JOIN car ON
      permit.car_id = car.id 
    WHERE
      permit.start_date <= NOW()
      AND permit.end_date >= NOW()
    LIMIT $1
    OFFSET $2
  `

	permits := []permit{}
	err := permitRepo.database.driver.Select(&permits, query, getBoundedLimit(limit), offset)
	if err != nil {
		return nil, fmt.Errorf("permit_repo: GetActive: %v", newError(ErrDatabaseQuery, err))
	}

	modelsPermits := slice.Map(permits, func(p permit) models.Permit { return p.toModels() })

	return modelsPermits, nil
}

func (permitRepo PermitRepo) GetAll(limit, offset uint) ([]models.Permit, error) {
	const query = `
    SELECT
      permit.id,
      permit.resident_id,
      car.id AS car_id,
      car.license_plate,
      car.color,
      car.make,
      car.model,
      permit.start_date,
      permit.end_date,
      permit.request_ts,
      permit.affects_days
    FROM permit
    LEFT JOIN car ON
      permit.car_id = car.id
    LIMIT $1
    OFFSET $2
  `

	permits := []permit{}
	err := permitRepo.database.driver.Select(&permits, query, getBoundedLimit(limit), offset)
	if err != nil {
		return nil, fmt.Errorf("permit_repo: GetAll: %v", newError(ErrDatabaseQuery, err))
	}

	modelsPermits := slice.Map(permits, func(p permit) models.Permit { return p.toModels() })

	return modelsPermits, nil
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
