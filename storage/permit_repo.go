package storage

import (
	"database/sql"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"strings"
	"time"
)

type PermitRepo struct {
	database Database
	selector Selector
}

func NewPermitRepo(database Database) PermitRepo {
	permitSelect := stmtBuilder.Select(
		"permit.id AS permit_id",
		"permit.resident_id",
		"permit.car_id",
		"permit.license_plate",
		"permit.color",
		"permit.make",
		"permit.model",
		"permit.start_ts",
		"permit.end_ts",
		"permit.request_ts",
		"permit.affects_days",
		"permit.exception_reason",
	).From("permit")
	countSelect := stmtBuilder.Select("count(*)").From("permit")

	selector := newSelector(permitSelect, countSelect).withOpts(withSearchFn(searchPermitsSQL))

	return PermitRepo{
		database: database,
		selector: selector,
	}
}

func (permitRepo PermitRepo) SelectWhere(permitFields models.Permit, selectOpts ...func(*Selector)) ([]models.Permit, error) {
	selector := permitRepo.selector.withOpts(selectOpts...)

	permitSelect := selector.selector.Where(rmEmptyVals(squirrel.Eq{
		"resident_id":   permitFields.ResidentID,
		"car_id":        permitFields.CarID,
		"license_plate": permitFields.LicensePlate,
		"color":         permitFields.Color,
		"make":          permitFields.Make,
		"model":         permitFields.Model,
	}))
	if !permitFields.StartDate.IsZero() {
		permitSelect = permitSelect.Where("start_ts = ?", permitFields.StartDate.Unix())
	}
	if !permitFields.EndDate.IsZero() {
		permitSelect = permitSelect.Where("end_ts = ?", permitFields.EndDate.Unix())
	}

	query, args, err := permitSelect.ToSql()
	if err != nil {
		return nil, fmt.Errorf("permit_repo.Get: %w: %v", errs.DBBuildingQuery, err)
	}

	permits := permitSlice{}
	err = permitRepo.database.driver.Select(&permits, query, args...)
	if err != nil {
		return nil, fmt.Errorf("permit_repo.Get: %w: %v. %s. %v", errs.DBQuery, err, query, args)
	}

	return permits.toModels(), nil
}

func (permitRepo PermitRepo) SelectCountWhere(permitFields models.Permit, selectOpts ...func(*Selector)) (int, error) {
	selector := permitRepo.selector.withOpts(selectOpts...)

	countSelect := selector.countSelect.Where(rmEmptyVals(squirrel.Eq{
		"resident_id":   permitFields.ResidentID,
		"car_id":        permitFields.CarID,
		"license_plate": permitFields.LicensePlate,
		"color":         permitFields.Color,
		"make":          permitFields.Make,
		"model":         permitFields.Model,
	}))
	if !permitFields.StartDate.IsZero() {
		countSelect = countSelect.Where("start_ts = ?", permitFields.StartDate.Unix())
	}
	if !permitFields.EndDate.IsZero() {
		countSelect = countSelect.Where("end_ts = ?", permitFields.EndDate.Unix())
	}

	query, args, err := countSelect.ToSql()
	if err != nil {
		return 0, fmt.Errorf("permit_repo.GetCount: %w: %v", errs.DBBuildingQuery, err)
	}

	var totalAmount int
	err = permitRepo.database.driver.Get(&totalAmount, query, args...)
	if err != nil {
		return 0, fmt.Errorf("permit_repo.GetCount: %w: %v", errs.DBQuery, err)
	}

	return totalAmount, nil
}

func (permitRepo PermitRepo) GetOne(id int) (models.Permit, error) {
	query, args, err := permitRepo.selector.selector.Where("permit.id = $1", id).ToSql()
	if err != nil {
		return models.Permit{}, fmt.Errorf("permit_repo.GetOne: %w: %v", errs.DBBuildingQuery, err)
	}

	permit := permit{}
	err = permitRepo.database.driver.Get(&permit, query, args...)
	if err == sql.ErrNoRows {
		return models.Permit{}, fmt.Errorf("permit_repo.GetOne: %w", errs.NewNotFound("permit"))
	} else if err != nil {
		return models.Permit{}, fmt.Errorf("permit_repo.GetOne: %w: %v", errs.DBQuery, err)
	}

	return permit.toModels(), nil
}

func (permitRepo PermitRepo) Create(desiredPermit models.Permit) (int, error) {
	// see whether exceptionReason is empty and convert appropriately
	nullableReason := sql.NullString{}
	if desiredPermit.ExceptionReason != "" {
		nullableReason = sql.NullString{String: desiredPermit.ExceptionReason, Valid: true}
	}

	query, args, err := stmtBuilder.
		Insert("permit").
		SetMap(squirrel.Eq{
			"resident_id":      desiredPermit.ResidentID,
			"car_id":           desiredPermit.CarID,
			"license_plate":    desiredPermit.LicensePlate,
			"color":            desiredPermit.Color,
			"make":             desiredPermit.Make,
			"model":            desiredPermit.Model,
			"start_ts":         desiredPermit.StartDate.Unix(),
			"end_ts":           desiredPermit.EndDate.Unix(),
			"request_ts":       time.Now().Unix(),
			"affects_days":     desiredPermit.AffectsDays,
			"exception_reason": nullableReason,
		}).
		Suffix("RETURNING permit.id").
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("permit_repo.Create: %w: %v", errs.DBBuildingQuery, err)
	}

	var permitID int
	err = permitRepo.database.driver.Get(&permitID, query, args...)
	if err != nil {
		return 0, fmt.Errorf("permit_repo.Create: %w: %v", errs.DBExec, err)
	}

	return permitID, nil
}

func (permitRepo PermitRepo) Delete(id int) error {
	const query = `DELETE FROM permit WHERE id = $1`

	res, err := permitRepo.database.driver.Exec(query, id)
	if err != nil {
		return fmt.Errorf("permit_repo.Delete: %w: %v", errs.DBExec, err)
	}

	if rowsAffected, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("permit_repo.Delete: %w: %v", errs.DBGetRowsAffected, err)
	} else if rowsAffected == 0 {
		return fmt.Errorf("permit_repo.Delete: %w", errs.NewNotFound("permit"))
	}

	return nil
}

func (permitRepo PermitRepo) Update(permitFields models.Permit) error {
	permitUpdate := stmtBuilder.Update("permit").SetMap(rmEmptyVals(squirrel.Eq{
		"license_plate": permitFields.LicensePlate,
		"color":         permitFields.Color,
		"make":          permitFields.Make,
		"model":         permitFields.Model,
	}))

	query, args, err := permitUpdate.Where("permit.id = ?", permitFields.ID).ToSql()
	if err != nil {
		return fmt.Errorf("permit_repo.Update: %w: %v", errs.DBBuildingQuery, err)
	}

	_, err = permitRepo.database.driver.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("permit_repo.Update: %w: %v", errs.DBExec, err)
	}

	return nil
}

// helpers
func searchPermitsSQL(query string) squirrel.Sqlizer {
	lcQuery := strings.ToLower(query)
	return squirrel.Or{
		squirrel.Expr("LOWER(CAST(permit.id AS TEXT)) = ?", lcQuery),
		squirrel.Expr("LOWER(permit.resident_id) = ?", lcQuery),
		squirrel.Expr("LOWER(permit.license_plate) = ?", lcQuery),
		squirrel.Expr("LOWER(permit.color) = ?", lcQuery),
		squirrel.Expr("LOWER(permit.make) = ?", lcQuery),
		squirrel.Expr("LOWER(permit.model) = ?", lcQuery),
	}
}
