package storage

import (
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"strings"
)

type ResidentRepo struct {
	database Database
	selector Selector
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
	countSelect := squirrel.Select("count(*)").From("resident")

	selector := newSelector(residentSelect, countSelect).withOpts(withSearchFn(searchResidentsSQL))

	return ResidentRepo{
		database: database,
		selector: selector,
	}
}

func (residentRepo ResidentRepo) SelectWhere(residentFields models.Resident, selectOpts ...func(*Selector)) ([]models.Resident, error) {
	selector := residentRepo.selector.withOpts(selectOpts...)

	residentSelect := selector.selector.Where(rmEmptyVals(squirrel.Eq{
		"resident_id": residentFields.ID,
		"first_name":  residentFields.FirstName,
		"last_name":   residentFields.LastName,
		"phone":       residentFields.Phone,
		"email":       residentFields.Email,
	}))

	query, args, err := residentSelect.ToSql()
	if err != nil {
		return nil, fmt.Errorf("resident_repo.SelectWhere: %w: %v", errs.DBBuildingQuery, err)
	}

	residents := residentSlice{}
	err = residentRepo.database.driver.Select(&residents, query, args...)
	if err != nil {
		return nil, fmt.Errorf("resident_repo.SelectWhere: %w: %v", errs.DBQuery, err)
	}

	return residents.toModels(), nil
}

func (residentRepo ResidentRepo) SelectCountWhere(residentFields models.Resident, selectOpts ...func(*Selector)) (int, error) {
	selector := residentRepo.selector.withOpts(selectOpts...)

	countSelect := selector.countSelect.Where(rmEmptyVals(squirrel.Eq{
		"resident_id": residentFields.ID,
		"first_name":  residentFields.FirstName,
		"last_name":   residentFields.LastName,
		"phone":       residentFields.Phone,
		"email":       residentFields.Email,
	}))

	query, args, err := countSelect.ToSql()
	if err != nil {
		return 0, fmt.Errorf("resident_repo.SelectCountWhere: %w: %v", errs.DBBuildingQuery, err)
	}

	var totalAmount int
	err = residentRepo.database.driver.Get(&totalAmount, query, args...)
	if err != nil {
		return 0, fmt.Errorf("resident_repo.SelectCountWhere: %w: %v", errs.DBQuery, err)
	}

	return totalAmount, nil
}

func (residentRepo ResidentRepo) AddToAmtParkingDaysUsed(id string, days int) error {
	const query = `
    UPDATE resident SET amt_parking_days_used = amt_parking_days_used + $1
    WHERE id = $2
  `

	_, err := residentRepo.database.driver.Exec(query, days, id)
	if err != nil {
		return fmt.Errorf("resident_repo.AddToAmtParkingDaysUsed: %w: %v", errs.DBExec, err)
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
		return fmt.Errorf("resident_repo.Create: %w: %v", errs.DBBuildingQuery, err)
	}

	_, err = residentRepo.database.driver.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("resident_repo.Create: %w: %v", errs.DBExec, err)
	}

	return nil
}

func (residentRepo ResidentRepo) Delete(residentID string) error {
	const query = `DELETE FROM resident WHERE id = $1`

	res, err := residentRepo.database.driver.Exec(query, residentID)
	if err != nil {
		return fmt.Errorf("resident_repo.Delete: %w: %v", errs.DBExec, err)
	}

	if rowsAffected, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("resident_repo.Delete: %w: %v", errs.DBGetRowsAffected, err)
	} else if rowsAffected == 0 {
		return fmt.Errorf("resident_repo.Delete: %w", errs.NewNotFound("resident"))
	}

	return nil
}

func (residentRepo ResidentRepo) Update(residentFields models.Resident) error {
	residentUpdate := stmtBuilder.Update("resident").SetMap(rmEmptyVals(squirrel.Eq{
		"first_name":            residentFields.FirstName,
		"last_name":             residentFields.LastName,
		"phone":                 residentFields.Phone,
		"email":                 residentFields.Email,
		"password":              residentFields.Password,
		"unlim_days":            *residentFields.UnlimDays,
		"amt_parking_days_used": *residentFields.AmtParkingDaysUsed,
		"token_version":         residentFields.TokenVersion,
	}))

	query, args, err := residentUpdate.Where("resident.id = ?", residentFields.ID).ToSql()
	if err != nil {
		return fmt.Errorf("resident_repo.Update: %w: %v", errs.DBBuildingQuery, err)
	}

	_, err = residentRepo.database.driver.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("resident_repo.Update: %w: %v", errs.DBExec, err)
	}

	return nil
}

// helpers
func searchResidentsSQL(query string) squirrel.Sqlizer {
	lcQuery := strings.ToLower(query)
	return squirrel.Or{
		squirrel.Expr("LOWER(resident.id) = ?", strings.ToLower(lcQuery)),
		squirrel.Expr("LOWER(resident.first_name) = ?", lcQuery),
		squirrel.Expr("LOWER(resident.last_name) = ?", lcQuery),
	}
}
