package storage

import (
	"database/sql"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"strings"
	"time"
)

type PermitRepo struct {
	database     Database
	permitSelect squirrel.SelectBuilder
	filterToSQL  map[models.PermitFilter]squirrel.Sqlizer
	permitASC    string
	permitDESC   string
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
		"permit.exception_reason",
	).From("permit").
		LeftJoin("car ON permit.car_id = car.id")

	filterToSQL := map[models.PermitFilter]squirrel.Sqlizer{
		models.ActivePermits: squirrel.And{
			squirrel.Expr("permit.start_ts <= extract(epoch from now())"),
			squirrel.Expr("permit.end_ts >= extract(epoch from now())"),
		},
		models.ExceptionPermits: squirrel.Expr("permit.exception_reason IS NOT NULL"),
		models.ExpiredPermits: squirrel.And{
			squirrel.Expr("permit.end_ts >= extract(epoch from (CURRENT_DATE - '1 DAY'::interval * $1))", config.DefaultExpiredWindow),
			squirrel.Expr("permit.end_ts <= extract(epoch from (CURRENT_DATE-2))"),
		},
	}

	permitASC := "permit.id ASC"
	permitDESC := "permit.id DESC"

	return PermitRepo{
		database:     database,
		permitSelect: permitSelect,
		filterToSQL:  filterToSQL,
		permitASC:    permitASC,
		permitDESC:   permitDESC,
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
		return nil, fmt.Errorf("permit_repo.Get: %w: limit or offset cannot be zero", ErrInvalidArg)
	}

	permitSelect := permitRepo.permitSelect
	if whereSQL, ok := permitRepo.filterToSQL[filter]; ok {
		permitSelect = permitSelect.Where(whereSQL)
	}

	if residentID != "" {
		permitSelect = permitSelect.Where("permit.resident_id = $1", residentID)
	}

	if search != "" {
		permitSelect = permitSelect.
			Where(squirrel.Or{
				squirrel.Expr("LOWER(CAST(permit.id AS TEXT)) = $1", strings.ToLower(search)),
				squirrel.Expr("LOWER(permit.resident_id) = $1"),
				squirrel.Expr("LOWER(car.license_plate) = $1"),
				squirrel.Expr("LOWER(car.color) = $1"),
				squirrel.Expr("LOWER(car.make) = $1"),
				squirrel.Expr("LOWER(car.model) = $1"),
			})
	}

	if !reversed {
		permitSelect = permitSelect.OrderBy(permitRepo.permitASC)
	} else {
		permitSelect = permitSelect.OrderBy(permitRepo.permitDESC)
	}

	query, args, err := permitSelect.
		Limit(uint64(getBoundedLimit(limit))).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("permit_repo.Get: %w: %v", ErrBuildingQuery, err)
	}

	permits := permitSlice{}
	err = permitRepo.database.driver.Select(&permits, query, args...)
	if err != nil {
		return nil, fmt.Errorf("permit_repo.Get: %w: %v", ErrDatabaseQuery, err)
	}

	return permits.toModels(), nil
}

func (permitRepo PermitRepo) GetCount(filter models.PermitFilter, residentID string) (int, error) {
	countSelect := squirrel.Select("count(*)").From("permit")
	if whereSQL, ok := permitRepo.filterToSQL[filter]; ok {
		countSelect = countSelect.Where(whereSQL)
	}

	if residentID != "" {
		countSelect = countSelect.Where("permit.resident_id = $1", residentID)
	}

	query, args, err := countSelect.ToSql()
	if err != nil {
		return 0, fmt.Errorf("permit_repo.GetCount: %w: %v", ErrBuildingQuery, err)
	}

	var totalAmount int
	err = permitRepo.database.driver.Get(&totalAmount, query, args...)
	if err != nil {
		return 0, fmt.Errorf("permit_repo.GetCount: %w: %v", ErrDatabaseQuery, err)
	}

	return totalAmount, nil
}

func (permitRepo PermitRepo) GetOne(id int) (models.Permit, error) {
	if id == 0 {
		return models.Permit{}, fmt.Errorf("permit_repo.GetOne: %w: Empty ID argument", ErrInvalidArg)
	}

	query, args, err := permitRepo.permitSelect.Where("permit.id = $1", id).ToSql()
	if err != nil {
		return models.Permit{}, fmt.Errorf("permit_repo.GetOne: %w: %v", ErrBuildingQuery, err)
	}

	permit := permit{}
	err = permitRepo.database.driver.Get(&permit, query, args...)
	if err == sql.ErrNoRows {
		return models.Permit{}, fmt.Errorf("permit_repo.GetOne: %w", ErrNoRows)
	} else if err != nil {
		return models.Permit{}, fmt.Errorf("permit_repo.GetOne: %w: %v", ErrDatabaseQuery, err)
	}

	return permit.toModels(), nil
}

func (permitRepo PermitRepo) Create(
	residentId,
	carId string,
	startTS,
	endTS int64,
	affectsDays bool,
	exceptionReason string,
) (int, error) {
	// see whether exceptionReason is empty and convert appropriately
	// assume everything else is already checked for emptyness
	nullableReason := sql.NullString{}
	if exceptionReason != "" {
		nullableReason = sql.NullString{String: exceptionReason, Valid: true}
	}

	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := sq.
		Insert("permit").
		SetMap(squirrel.Eq{
			"resident_id":      residentId,
			"car_id":           carId,
			"start_ts":         startTS,
			"end_ts":           endTS,
			"request_ts":       time.Now().Unix(),
			"affects_days":     affectsDays,
			"exception_reason": nullableReason,
		}).
		Suffix("RETURNING permit.id").
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("permit_repo.Create: %w: %v", ErrBuildingQuery, err)
	}

	var permitId int
	err = permitRepo.database.driver.Get(&permitId, query, args...)
	if err != nil {
		return 0, fmt.Errorf("permit_repo.Create: %w: %v", ErrDatabaseExec, err)
	}

	return permitId, nil
}

func (permitRepo PermitRepo) GetActiveOfCarDuring(carId string, startDate, endDate int64) ([]models.Permit, error) {
	query, args, err := permitRepo.permitSelect.
		Where("car_id = $1", carId).
		Where("permit.start_ts <= $2", endDate).
		Where("permit.end_ts >= $3", startDate).
		OrderBy(permitRepo.permitASC).
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

func (permitRepo PermitRepo) GetActiveOfResidentDuring(residentId string, startDate, endDate int64) ([]models.Permit, error) {
	if residentId == "" {
		return []models.Permit{}, fmt.Errorf("permit_repo.GetActiveOfResidentDuring: %w: Empty ID argument", ErrInvalidArg)
	}

	query, args, err := permitRepo.permitSelect.
		Where("permit.resident_id = $1", residentId).
		Where("permit.start_ts <= $2", endDate).
		Where("permit.end_ts >= $3", startDate).
		OrderBy(permitRepo.permitASC).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("permit_repo.GetActiveOfResidentDuring: %w: %v", ErrBuildingQuery, err)
	}

	permits := permitSlice{}
	err = permitRepo.database.driver.Select(&permits, query, args...)
	if err != nil {
		return nil, fmt.Errorf("permit_repo.GetActiveOfResidentDuring: %w: %v", ErrDatabaseQuery, err)
	}

	return permits.toModels(), nil
}

func (permitRepo PermitRepo) Delete(id int) error {
	if id <= 0 {
		return fmt.Errorf("permit_repo.Delete: %w: negative or zero ID argument", ErrInvalidArg)
	}
	const query = `DELETE FROM permit WHERE id = $1`

	res, err := permitRepo.database.driver.Exec(query, id)
	if err != nil {
		return fmt.Errorf("permit_repo.Delete: %w: %v", ErrDatabaseExec, err)
	}

	if rowsAffected, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("permit_repo.Delete: %w: %v", ErrGetRowsAffected, err)
	} else if rowsAffected == 0 {
		return fmt.Errorf("permit_repo.Delete: %w", ErrNoRows)
	}

	return nil
}
