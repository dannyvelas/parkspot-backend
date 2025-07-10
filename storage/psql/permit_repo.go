package psql

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/parkspot-backend/errs"
	"github.com/dannyvelas/parkspot-backend/models"
	"github.com/dannyvelas/parkspot-backend/storage"
	"github.com/dannyvelas/parkspot-backend/storage/selectopts"
	"github.com/jmoiron/sqlx"
)

type PermitRepo struct {
	driver       *sqlx.DB
	permitSelect squirrel.SelectBuilder
	countSelect  squirrel.SelectBuilder
}

func NewPermitRepo(driver *sqlx.DB) storage.PermitRepo {
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

	return PermitRepo{
		driver:       driver,
		permitSelect: permitSelect,
		countSelect:  countSelect,
	}
}

func (permitRepo PermitRepo) SelectWhere(permitFields models.Permit, selectOpts ...selectopts.SelectOpt) ([]models.Permit, error) {
	selector := permitRepo.permitSelect
	for _, opt := range selectOpts {
		selector = opt.Dispatch(permitRepo, selector)
	}

	permitSelect := selector.Where(rmEmptyVals(squirrel.Eq{
		"resident_id":   permitFields.ResidentID,
		"car_id":        permitFields.CarID,
		"license_plate": permitFields.LicensePlate,
		"color":         permitFields.Color,
		"make":          permitFields.Make,
		"model":         permitFields.Model,
	}))

	query, args, err := permitSelect.ToSql()
	if err != nil {
		return nil, fmt.Errorf("permit_repo.SelectWhere: %w: %v", errs.DBBuildingQuery, err)
	}

	permits := permitSlice{}
	err = permitRepo.driver.Select(&permits, query, args...)
	if err != nil {
		return nil, fmt.Errorf("permit_repo.SelectWhere: %w: %v. %s. %v", errs.DBQuery, err, query, args)
	}

	return permits.toModels(), nil
}

func (permitRepo PermitRepo) SelectCountWhere(permitFields models.Permit, selectOpts ...selectopts.SelectOpt) (int, error) {
	selector := permitRepo.countSelect
	for _, opt := range selectOpts {
		selector = opt.Dispatch(permitRepo, selector)
	}

	countSelect := selector.Where(rmEmptyVals(squirrel.Eq{
		"resident_id":   permitFields.ResidentID,
		"car_id":        permitFields.CarID,
		"license_plate": permitFields.LicensePlate,
		"color":         permitFields.Color,
		"make":          permitFields.Make,
		"model":         permitFields.Model,
	}))

	query, args, err := countSelect.ToSql()
	if err != nil {
		return 0, fmt.Errorf("permit_repo.GetCount: %w: %v", errs.DBBuildingQuery, err)
	}

	var totalAmount int
	err = permitRepo.driver.Get(&totalAmount, query, args...)
	if err != nil {
		return 0, fmt.Errorf("permit_repo.GetCount: %w: %v", errs.DBQuery, err)
	}

	return totalAmount, nil
}

func (permitRepo PermitRepo) GetOne(id int) (models.Permit, error) {
	query, args, err := permitRepo.permitSelect.Where("permit.id = $1", id).ToSql()
	if err != nil {
		return models.Permit{}, fmt.Errorf("permit_repo.GetOne: %w: %v", errs.DBBuildingQuery, err)
	}

	permit := permit{}
	err = permitRepo.driver.Get(&permit, query, args...)
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
	err = permitRepo.driver.Get(&permitID, query, args...)
	if err != nil {
		return 0, fmt.Errorf("permit_repo.Create: %w: %v", errs.DBExec, err)
	}

	return permitID, nil
}

func (permitRepo PermitRepo) Delete(id int) error {
	const query = `DELETE FROM permit WHERE id = $1`

	res, err := permitRepo.driver.Exec(query, id)
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

	_, err = permitRepo.driver.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("permit_repo.Update: %w: %v", errs.DBExec, err)
	}

	return nil
}

func (permitRepo PermitRepo) Reset() error {
	_, err := permitRepo.driver.Exec("DELETE FROM permit")
	if err != nil {
		return fmt.Errorf("permit_repo.Reset: %w: %v", errs.DBExec, err)
	}

	return nil
}

// helpers
func (permitRepo PermitRepo) SearchAsSQL(query string) squirrel.Sqlizer {
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

func (permitRepo PermitRepo) StatusAsSQL(status models.Status) (squirrel.Sqlizer, bool) {
	statusToSQL := map[models.Status]squirrel.Sqlizer{
		models.ActiveStatus: squirrel.And{
			squirrel.Expr("permit.start_ts <= extract(epoch from now())"),
			squirrel.Expr("permit.end_ts >= extract(epoch from now())"),
		},
		models.ExceptionStatus: squirrel.Expr("permit.exception_reason IS NOT NULL"),
		models.ExpiredStatus:   squirrel.Expr("permit.end_ts <= extract(epoch from (CURRENT_DATE-2))"),
	}

	whereSQL, ok := statusToSQL[status]
	return whereSQL, ok
}
