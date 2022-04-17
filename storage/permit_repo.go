package storage

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/typesafe"
)

type PermitRepo struct {
	database Database
	carRepo  carRepo
}

func NewPermitRepo(database Database) PermitRepo {
	return PermitRepo{database: database, carRepo: newCarRepo(database)}
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
		return nil, fmt.Errorf("permit_repo: GetActive: %w", newError(ErrDatabaseQuery, err))
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
		return nil, fmt.Errorf("permit_repo: GetAll: %w", newError(ErrDatabaseQuery, err))
	}

	return permits.toModels(), nil
}

func (permitRepo PermitRepo) Create(permit models.Permit) (models.Permit, error) {
	zeroValFields := typesafe.ZeroValFields(permit)
	if len(zeroValFields) > 0 {
		return models.Permit{}, fmt.Errorf("permit_repo: Create: %w", errMissingFields(zeroValFields))
	}

	car, err := permitRepo.carRepo.CreateIfNotExists(permit.Car)
	if err != nil {
		return models.Permit{}, fmt.Errorf("permit_repo: Create: %w", err)
	}

	const permitQuery = `
    INSERT INTO permit(id, resident_id, car_id, start_ts, end_ts, request_ts, affects_days)
    VALUES($1, $2, $3, $4, $5, $6, $7);
  `

	_, err = permitRepo.database.driver.Exec(permitQuery, permit.Id, permit.ResidentId, car.Id,
		permit.StartDate.Unix(), permit.EndDate.Unix(), permit.RequestTS, permit.AffectsDays)
	if err != nil {
		return models.Permit{}, fmt.Errorf("permit_repo: Create: %w", newError(ErrDatabaseExec, err))
	}

	return permit, nil
}
