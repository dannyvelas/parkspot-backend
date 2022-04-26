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
