package psql

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/parkspot-backend/models"
	"github.com/dannyvelas/parkspot-backend/storage"
	"github.com/dannyvelas/parkspot-backend/storage/selectopts"
	"github.com/jmoiron/sqlx"
)

type ResidentRepo struct {
	driver         *sqlx.DB
	residentSelect squirrel.SelectBuilder
	countSelect    squirrel.SelectBuilder
}

func NewResidentRepo(driver *sqlx.DB) storage.ResidentRepo {
	residentSelect := stmtBuilder.Select(
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
	countSelect := stmtBuilder.Select("count(*)").From("resident")

	return ResidentRepo{
		driver:         driver,
		residentSelect: residentSelect,
		countSelect:    countSelect,
	}
}

func (residentRepo ResidentRepo) GetOne(id string) (models.Resident, error) {
	query, args, err := residentRepo.residentSelect.Where("id = $1", id).ToSql()
	if err != nil {
		return models.Resident{}, fmt.Errorf("resident_repo.GetOne: %w: %v", storage.ErrDBBuildingQuery, err)
	}

	resident := resident{}
	err = residentRepo.driver.Get(&resident, query, args...)
	if err == sql.ErrNoRows {
		return models.Resident{}, fmt.Errorf("resident_repo.GetOne: %w", storage.NewNotFound("resident"))
	} else if err != nil {
		return models.Resident{}, fmt.Errorf("resident_repo.GetOne: %w: %v", storage.ErrDBQuery, err)
	}

	return resident.toModels(), nil
}

func (residentRepo ResidentRepo) SelectWhere(residentFields models.Resident, selectOpts ...selectopts.SelectOpt) ([]models.Resident, error) {
	selector := residentRepo.residentSelect
	for _, opt := range selectOpts {
		selector = opt.Dispatch(residentRepo, selector)
	}

	residentSelect := selector.Where(rmEmptyVals(squirrel.Eq{
		"id":         residentFields.ID,
		"first_name": residentFields.FirstName,
		"last_name":  residentFields.LastName,
		"phone":      residentFields.Phone,
		"email":      residentFields.Email,
	})).OrderBy("first_name ASC")

	query, args, err := residentSelect.ToSql()
	if err != nil {
		return nil, fmt.Errorf("resident_repo.SelectWhere: %w: %v", storage.ErrDBBuildingQuery, err)
	}

	residents := residentSlice{}
	err = residentRepo.driver.Select(&residents, query, args...)
	if err != nil {
		return nil, fmt.Errorf("resident_repo.SelectWhere: %w: %v", storage.ErrDBQuery, err)
	}

	return residents.toModels(), nil
}

func (residentRepo ResidentRepo) SelectCountWhere(residentFields models.Resident, selectOpts ...selectopts.SelectOpt) (int, error) {
	selector := residentRepo.countSelect
	for _, opt := range selectOpts {
		selector = opt.Dispatch(residentRepo, selector)
	}

	countSelect := selector.Where(rmEmptyVals(squirrel.Eq{
		"id":         residentFields.ID,
		"first_name": residentFields.FirstName,
		"last_name":  residentFields.LastName,
		"phone":      residentFields.Phone,
		"email":      residentFields.Email,
	}))

	query, args, err := countSelect.ToSql()
	if err != nil {
		return 0, fmt.Errorf("resident_repo.SelectCountWhere: %w: %v", storage.ErrDBBuildingQuery, err)
	}

	var totalAmount int
	err = residentRepo.driver.Get(&totalAmount, query, args...)
	if err != nil {
		return 0, fmt.Errorf("resident_repo.SelectCountWhere: %w: %v", storage.ErrDBQuery, err)
	}

	return totalAmount, nil
}

func (residentRepo ResidentRepo) AddToAmtParkingDaysUsed(id string, days int) error {
	const query = `
    UPDATE resident SET amt_parking_days_used = amt_parking_days_used + $1
    WHERE id = $2
  `

	_, err := residentRepo.driver.Exec(query, days, id)
	if err != nil {
		return fmt.Errorf("resident_repo.AddToAmtParkingDaysUsed: %w: %v", storage.ErrDBExec, err)
	}

	return nil
}

func (residentRepo ResidentRepo) Create(resident models.Resident) error {
	// cast *resident.UnlimDays to bool
	unlimDays := false
	if resident.UnlimDays != nil {
		unlimDays = *resident.UnlimDays
	}

	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := sq.
		Insert("resident").
		SetMap(squirrel.Eq{
			"id":         resident.ID,
			"first_name": resident.FirstName,
			"last_name":  resident.LastName,
			"phone":      resident.Phone,
			"email":      resident.Email,
			"password":   resident.Password,
			"unlim_days": unlimDays,
		}).ToSql()
	if err != nil {
		return fmt.Errorf("resident_repo.Create: %w: %v", storage.ErrDBBuildingQuery, err)
	}

	_, err = residentRepo.driver.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("resident_repo.Create: %w: %v", storage.ErrDBExec, err)
	}

	return nil
}

func (residentRepo ResidentRepo) Delete(residentID string) error {
	const query = `DELETE FROM resident WHERE id = $1`

	res, err := residentRepo.driver.Exec(query, residentID)
	if err != nil {
		return fmt.Errorf("resident_repo.Delete: %w: %v", storage.ErrDBExec, err)
	}

	if rowsAffected, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("resident_repo.Delete: %w: %v", storage.ErrDBGetRowsAffected, err)
	} else if rowsAffected == 0 {
		return fmt.Errorf("resident_repo.Delete: %w", storage.NewNotFound("resident"))
	}

	return nil
}

func (residentRepo ResidentRepo) Update(residentFields models.Resident) error {
	residentUpdate := stmtBuilder.Update("resident").SetMap(rmEmptyVals(squirrel.Eq{
		"first_name": residentFields.FirstName,
		"last_name":  residentFields.LastName,
		"phone":      residentFields.Phone,
		"email":      residentFields.Email,
		"password":   residentFields.Password,
	}))
	if residentFields.UnlimDays != nil {
		residentUpdate = residentUpdate.Set("unlim_days", *residentFields.UnlimDays)
	}
	if residentFields.AmtParkingDaysUsed != nil {
		residentUpdate = residentUpdate.Set("amt_parking_days_used", *residentFields.AmtParkingDaysUsed)
	}
	if residentFields.TokenVersion != nil {
		residentUpdate = residentUpdate.Set("token_version", *residentFields.TokenVersion)
	}

	query, args, err := residentUpdate.Where("resident.id = ?", residentFields.ID).ToSql()
	if err != nil {
		return fmt.Errorf("resident_repo.Update: %w: %v", storage.ErrDBBuildingQuery, err)
	}

	_, err = residentRepo.driver.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("resident_repo.Update: %w: %v", storage.ErrDBExec, err)
	}

	return nil
}

func (residentRepo ResidentRepo) SearchAsSQL(query string) squirrel.Sqlizer {
	lcQuery := strings.ToLower(query)
	return squirrel.Or{
		squirrel.Expr("LOWER(resident.id) = ?", strings.ToLower(lcQuery)),
		squirrel.Expr("LOWER(resident.first_name) = ?", lcQuery),
		squirrel.Expr("LOWER(resident.last_name) = ?", lcQuery),
	}
}
