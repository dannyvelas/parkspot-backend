package storage

import (
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/lasvistas_api/models"
	"time"
)

type PermitRepo struct {
	database     Database
	permitSelect squirrel.SelectBuilder
}

func NewPermitRepo(database Database) PermitRepo {
	permitSelect := squirrel.Select(
		"permit.id AS permit_id",
		"permit.resident_id",
		"car.id AS car_id",
		"car.license_plate",
		"car.color",
		"car.make",
		"car.model",
		"permit.start_ts",
		"permit.end_ts",
		"permit.request_ts",
		"permit.affects_days",
	).From("permit").LeftJoin("car ON permit.car_id = car.id")

	return PermitRepo{database: database, permitSelect: permitSelect}
}

func (permitRepo PermitRepo) GetActive(limit, offset uint64) ([]models.Permit, error) {
	query, _, err := permitRepo.permitSelect.
		Where("permit.start_ts <= extract(epoch from now())").
		Where("permit.end_ts >= extract(epoch from now())").
		Limit(getBoundedLimit(limit)).
		Offset(offset).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("permit_repo.GetActive: %w: %v", ErrBuildingQuery, err)
	}

	permits := permitSlice{}
	err = permitRepo.database.driver.Select(&permits, query)
	if err != nil {
		return nil, fmt.Errorf("permit_repo.GetActive: %w: %v", ErrDatabaseQuery, err)
	}

	return permits.toModels(), nil
}

func (permitRepo PermitRepo) GetAll(limit, offset uint64) ([]models.Permit, error) {
	query, _, err := permitRepo.permitSelect.Limit(getBoundedLimit(limit)).Offset(offset).ToSql()
	if err != nil {
		return nil, fmt.Errorf("permit_repo.GetAll: %w: %v", ErrBuildingQuery, err)
	}

	permits := permitSlice{}
	err = permitRepo.database.driver.Select(&permits, query)
	if err != nil {
		return nil, fmt.Errorf("permit_repo.GetAll: %w: %v", ErrDatabaseQuery, err)
	}

	return permits.toModels(), nil
}

func (permitRepo PermitRepo) Create(permit models.Permit) (models.Permit, error) {
	if err := permit.Validate(); err != nil {
		return models.Permit{}, fmt.Errorf("permit_repo.Create: %w", err)
	}

	const query = `
    INSERT INTO permit(id, resident_id, car_id, start_ts, end_ts, request_ts, affects_days)
    VALUES($1, $2, $3, $4, $5, $6, $7);
  `

	_, err := permitRepo.database.driver.Exec(query, permit.Id, permit.ResidentId, permit.Car.Id,
		permit.StartDate.Unix(), permit.EndDate.Unix(), permit.RequestTS, permit.AffectsDays)
	if err != nil {
		return models.Permit{}, fmt.Errorf("permit_repo.Create: %w: %v", ErrDatabaseExec, err)
	}

	return permit, nil
}

func (permitRepo PermitRepo) GetActiveOfCarDuring(carId string, startDate, endDate time.Time) ([]models.Permit, error) {
	query, args, err := permitRepo.permitSelect.
		Where("car_id = $1", carId).
		Where("permit.start_ts <= $2", endDate.Unix()).
		Where("permit.end_ts >= $3", startDate.Unix()).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("permit_repo.GetActiveOfCarDuring: %w: %v", ErrBuildingQuery, err)
	}

	permits := permitSlice{}
	err = permitRepo.database.driver.Select(&permits, query, args...)
	if err != nil {
		return nil, fmt.Errorf("permit_repo.GetActiveOfCarDuring: %w: %v", ErrDatabaseQuery, err)
	}

	return permits.toModels(), nil
}
