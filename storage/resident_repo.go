package storage

import (
	"database/sql"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/lasvistas_api/models"
)

type ResidentRepo struct {
	database       Database
	residentSelect squirrel.SelectBuilder
}

func NewResidentRepo(database Database) ResidentRepo {
	residentSelect := squirrel.Select(
		"id",
		"first_name",
		"last_name",
		"phone",
		"email",
		"password",
		"unlim_days",
		"amt_parking_days_used",
	).From("resident")

	return ResidentRepo{database: database, residentSelect: residentSelect}
}

func (residentRepo ResidentRepo) GetOne(id string) (models.Resident, error) {
	if id == "" {
		return models.Resident{}, fmt.Errorf("resident_repo.GetOne: %w: Empty ID argument", ErrInvalidArg)
	}

	query, args, err := residentRepo.residentSelect.Where("resident.id = $1", id).ToSql()
	if err != nil {
		return models.Resident{}, fmt.Errorf("resident_repo.GetOne: %w: %v", ErrBuildingQuery, err)
	}

	resident := resident{}
	err = residentRepo.database.driver.Get(&resident, query, args...)
	if err == sql.ErrNoRows {
		return models.Resident{}, fmt.Errorf("resident_repo.GetOne: %w", ErrNoRows)
	} else if err != nil {
		return models.Resident{}, fmt.Errorf("resident_repo.GetOne: %w: %v", ErrDatabaseQuery, err)
	}

	return resident.toModels(), nil
}

func (residentRepo ResidentRepo) GetAll(limit, offset int) ([]models.Resident, error) {
	if limit < 0 || offset < 0 {
		return nil, fmt.Errorf("permit_repo.GetActive: %w: limit or offset cannot be zero", ErrInvalidArg)
	}

	query, _, err := residentRepo.residentSelect.
		Limit(uint64(getBoundedLimit(limit))).
		Offset(uint64(offset)).
		OrderBy("resident.id ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("resident_repo.GetAll: %w: %v", ErrBuildingQuery, err)
	}

	residents := residentSlice{}
	err = residentRepo.database.driver.Select(&residents, query)
	if err != nil {
		return nil, fmt.Errorf("resident_repo.GetAll: %w: %v", ErrDatabaseQuery, err)
	}

	return residents.toModels(), nil
}

func (residentRepo ResidentRepo) AddToAmtParkingDaysUsed(id string, days int) error {
	const query = `
    UPDATE resident SET amt_parking_days_used = amt_parking_days_used + $1
    WHERE id = $2
  `

	_, err := residentRepo.database.driver.Exec(query, days, id)
	if err != nil {
		return fmt.Errorf("resident_repo.AddToAmtParkingDaysUsed: %w: %v", ErrDatabaseExec, err)
	}

	return nil
}
