package storage

import (
	"database/sql"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"strings"
	"time"
)

var (
	permitSelect = stmtBuilder.Select(
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

	filterToSQL = map[models.PermitFilter]squirrel.Sqlizer{
		models.ActivePermits: squirrel.And{
			squirrel.Expr("permit.start_ts <= extract(epoch from now())"),
			squirrel.Expr("permit.end_ts >= extract(epoch from now())"),
		},
		models.ExceptionPermits: squirrel.Expr("permit.exception_reason IS NOT NULL"),
		models.ExpiredPermits: squirrel.And{
			squirrel.Expr("permit.end_ts >= extract(epoch from (CURRENT_DATE - '1 DAY'::interval * ?))", config.DefaultExpiredWindow),
			squirrel.Expr("permit.end_ts <= extract(epoch from (CURRENT_DATE-2))"),
		},
	}

	permitASC  = "permit.id ASC"
	permitDESC = "permit.id DESC"
)

type PermitRepo struct {
	database Database
}

func NewPermitRepo(database Database) PermitRepo {
	return PermitRepo{
		database: database,
	}
}

func (permitRepo PermitRepo) Get(
	filter models.PermitFilter,
	residentID string,
	limit,
	offset int,
	reversed bool,
	search string,
) ([]models.Permit, error) {
	if limit < 0 || offset < 0 {
		return nil, fmt.Errorf("permit_repo.Get: %w: limit or offset cannot be smaller than zero", errs.DBInvalidArg)
	}

	permitSelect := permitSelect // deep copy global variable to avoid mutating it
	if whereSQL, ok := filterToSQL[filter]; ok {
		permitSelect = permitSelect.Where(whereSQL)
	}

	if residentID != "" {
		permitSelect = permitSelect.Where("permit.resident_id = ?", residentID)
	}

	if search != "" {
		permitSelect = permitSelect.Where(permitRepo.cellEquals(search))
	}

	if !reversed {
		permitSelect = permitSelect.OrderBy(permitASC)
	} else {
		permitSelect = permitSelect.OrderBy(permitDESC)
	}

	query, args, err := permitSelect.
		Limit(uint64(getBoundedLimit(limit))).
		Offset(uint64(offset)).
		ToSql()
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

func (permitRepo PermitRepo) GetCount(filter models.PermitFilter, residentID, search string) (int, error) {
	countSelect := stmtBuilder.Select("count(*)").From("permit")
	if whereSQL, ok := filterToSQL[filter]; ok {
		countSelect = countSelect.Where(whereSQL)
	}

	if residentID != "" {
		countSelect = countSelect.Where("permit.resident_id = ?", residentID)
	}

	if search != "" {
		countSelect = countSelect.Where(permitRepo.cellEquals(search))
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
	if id == 0 {
		return models.Permit{}, fmt.Errorf("permit_repo.GetOne: %w: Empty ID argument", errs.DBInvalidArg)
	}

	query, args, err := permitSelect.Where("permit.id = $1", id).ToSql()
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
	// assume everything else is already checked for emptyness
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

func (permitRepo PermitRepo) GetActiveOfCarDuring(carID string, startDate, endDate time.Time) ([]models.Permit, error) {
	query, args, err := permitSelect.
		Where("car_id = $1", carID).
		Where("permit.start_ts <= $2", endDate.Unix()).
		Where("permit.end_ts >= $3", startDate.Unix()).
		OrderBy(permitASC).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("permit_repo.GetActiveOfCarDuring: %w: %v", errs.DBBuildingQuery, err)
	}

	permits := permitSlice{}
	err = permitRepo.database.driver.Select(&permits, query, args...)
	if err != nil {
		return nil, fmt.Errorf("permit_repo.GetActiveOfCarDuring: %w: %v", errs.DBQuery, err)
	}

	return permits.toModels(), nil
}

func (permitRepo PermitRepo) GetActiveOfResidentDuring(residentID string, startDate, endDate time.Time) ([]models.Permit, error) {
	if residentID == "" {
		return nil, fmt.Errorf("permit_repo.GetActiveOfResidentDuring: %w: Empty ID argument", errs.DBInvalidArg)
	}

	query, args, err := permitSelect.
		Where("permit.resident_id = $1", residentID).
		Where("permit.start_ts <= $2", endDate.Unix()).
		Where("permit.end_ts >= $3", startDate.Unix()).
		OrderBy(permitASC).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("permit_repo.GetActiveOfResidentDuring: %w: %v", errs.DBBuildingQuery, err)
	}

	permits := permitSlice{}
	err = permitRepo.database.driver.Select(&permits, query, args...)
	if err != nil {
		return nil, fmt.Errorf("permit_repo.GetActiveOfResidentDuring: %w: %v", errs.DBQuery, err)
	}

	return permits.toModels(), nil
}

func (permitRepo PermitRepo) Delete(id int) error {
	if id <= 0 {
		return fmt.Errorf("permit_repo.Delete: %w: negative or zero ID argument", errs.DBInvalidArg)
	}
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

func (permitRepo PermitRepo) Update(editPermit models.Permit) error {
	permitUpdate := stmtBuilder.Update("permit")

	if editPermit.LicensePlate != "" {
		permitUpdate = permitUpdate.Set("license_plate", editPermit.LicensePlate)
	}
	if editPermit.Color != "" {
		permitUpdate = permitUpdate.Set("color", editPermit.Color)
	}
	if editPermit.Make != "" {
		permitUpdate = permitUpdate.Set("make", editPermit.Make)
	}
	if editPermit.Model != "" {
		permitUpdate = permitUpdate.Set("model", editPermit.Model)
	}

	query, args, err := permitUpdate.Where("permit.id = ?", editPermit.ID).ToSql()
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
func (permitRepo PermitRepo) cellEquals(query string) squirrel.Sqlizer {
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
