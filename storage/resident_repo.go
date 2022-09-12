package storage

import (
	"database/sql"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/lasvistas_api/models"
	"strings"
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
		"token_version",
	).From("resident")

	return ResidentRepo{database: database, residentSelect: residentSelect}
}

func (residentRepo ResidentRepo) GetOne(residentId string) (models.Resident, error) {
	if residentId == "" {
		return models.Resident{}, fmt.Errorf("resident_repo.GetOne: %w: Empty ID argument", ErrInvalidArg)
	}

	query, args, err := residentRepo.residentSelect.Where("resident.id = $1", residentId).ToSql()
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

func (residentRepo ResidentRepo) GetOneByEmail(email string) (models.Resident, error) {
	if email == "" {
		return models.Resident{}, fmt.Errorf("resident_repo.GetOneByEmail: %w: Empty ID argument", ErrInvalidArg)
	}

	query, args, err := residentRepo.residentSelect.Where("resident.email = $1", email).ToSql()
	if err != nil {
		return models.Resident{}, fmt.Errorf("resident_repo.GetOneByEmail: %w: %v", ErrBuildingQuery, err)
	}

	resident := resident{}
	err = residentRepo.database.driver.Get(&resident, query, args...)
	if err == sql.ErrNoRows {
		return models.Resident{}, fmt.Errorf("resident_repo.GetOneByEmail: %w", ErrNoRows)
	} else if err != nil {
		return models.Resident{}, fmt.Errorf("resident_repo.GetOneByEmail: %w: %v", ErrDatabaseQuery, err)
	}

	return resident.toModels(), nil
}

func (residentRepo ResidentRepo) GetAll(limit, offset int, search string) ([]models.Resident, error) {
	if limit < 0 || offset < 0 {
		return nil, fmt.Errorf("resident_repo.GetAll: %w: limit or offset cannot be zero", ErrInvalidArg)
	}

	residentSelect := residentRepo.residentSelect
	if search != "" {
		residentSelect = residentSelect.
			Where(squirrel.Or{
				squirrel.Expr("LOWER(resident.id) = $1", strings.ToLower(search)),
				squirrel.Expr("LOWER(resident.first_name) = $1"),
				squirrel.Expr("LOWER(resident.last_name) = $1"),
			})
	}

	query, args, err := residentSelect.
		Limit(uint64(getBoundedLimit(limit))).
		Offset(uint64(offset)).
		OrderBy("resident.id ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("resident_repo.GetAll: %w: %v", ErrBuildingQuery, err)
	}

	residents := residentSlice{}
	err = residentRepo.database.driver.Select(&residents, query, args...)
	if err != nil {
		return nil, fmt.Errorf("resident_repo.GetAll: %w: %v", ErrDatabaseQuery, err)
	}

	return residents.toModels(), nil
}

func (residentRepo ResidentRepo) GetAllTotalAmount() (int, error) {
	const query = "SELECT count(*) FROM resident"

	var totalAmount int
	err := residentRepo.database.driver.Get(&totalAmount, query)
	if err != nil {
		return 0, fmt.Errorf("resident_repo.GetAllTotalAmount: %w: %v", ErrDatabaseQuery, err)
	}

	return totalAmount, nil
}

func (residentRepo ResidentRepo) AddToAmtParkingDaysUsed(id string, days int) error {
	if id == "" {
		return fmt.Errorf("resident_repo.AddToAmtParkingDaysUsed: %w: Empty ID argument", ErrInvalidArg)
	}

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

func (residentRepo ResidentRepo) SetPasswordFor(id string, password string) error {
	if id == "" {
		return fmt.Errorf("resident_repo.GetOne: %w: Empty ID argument", ErrInvalidArg)
	} else if password == "" {
		return fmt.Errorf("resident_repo.GetOne: %w: Emtpy Password argument", ErrInvalidArg)
	}

	const query = `UPDATE resident SET password = $1 WHERE id = $2`
	_, err := residentRepo.database.driver.Exec(query, password, id)
	if err != nil {
		return fmt.Errorf("resident_repo.SetPasswordFor: %w: %v", ErrDatabaseExec, err)
	}

	return nil
}

func (residentRepo ResidentRepo) Create(residentId, firstName, lastName, phone, email, hash string, unlimDays bool) error {
	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := sq.
		Insert("resident").
		SetMap(squirrel.Eq{
			"id":         residentId,
			"first_name": firstName,
			"last_name":  lastName,
			"phone":      phone,
			"email":      email,
			"password":   hash,
			"unlim_days": unlimDays,
		}).ToSql()
	if err != nil {
		return fmt.Errorf("resident_repo.Create: %w: %v", ErrBuildingQuery, err)
	}

	_, err = residentRepo.database.driver.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("resident_repo.Create: %w: %v", ErrDatabaseExec, err)
	}

	return nil
}

func (residentRepo ResidentRepo) Delete(residentId string) error {
	if residentId == "" {
		return fmt.Errorf("resident_repo.Delete: %w: negative or zero ID argument", ErrInvalidArg)
	}
	const query = `DELETE FROM resident WHERE id = $1`

	res, err := residentRepo.database.driver.Exec(query, residentId)
	if err != nil {
		return fmt.Errorf("resident_repo.Delete: %w: %v", ErrDatabaseExec, err)
	}

	if rowsAffected, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("resident_repo.Delete: %w: %v", ErrGetRowsAffected, err)
	} else if rowsAffected == 0 {
		return fmt.Errorf("resident_repo.Delete: %w", ErrNoRows)
	}

	return nil
}
