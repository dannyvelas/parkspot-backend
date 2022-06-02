package storage

import (
	"database/sql"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/lasvistas_api/models"
	"time"
)

type PermitRepo struct {
	database       Database
	permitSelect   squirrel.SelectBuilder
	countSelect    squirrel.SelectBuilder
	activeWhere    squirrel.And
	exceptionWhere squirrel.Sqlizer
	expiredWhere   func(int) squirrel.And
	permitASC      string
	permitDESC     string
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

	countSelect := squirrel.Select("count(*)").From("permit")

	activeWhere := squirrel.And{
		squirrel.Expr("permit.start_ts <= extract(epoch from now())"),
		squirrel.Expr("permit.end_ts >= extract(epoch from now())"),
	}

	exceptionWhere := squirrel.Expr("permit.exception_reason IS NOT NULL")

	expiredWhere := func(window int) squirrel.And {
		return squirrel.And{
			squirrel.Expr("permit.end_ts >= extract(epoch from (CURRENT_DATE - '1 DAY'::interval * $1))", window),
			squirrel.Expr("permit.end_ts <= extract(epoch from (CURRENT_DATE-2))"),
		}
	}

	permitASC := "permit.id ASC"
	permitDESC := "permit.id DESC"

	return PermitRepo{
		database:       database,
		permitSelect:   permitSelect,
		countSelect:    countSelect,
		activeWhere:    activeWhere,
		exceptionWhere: exceptionWhere,
		expiredWhere:   expiredWhere,
		permitASC:      permitASC,
		permitDESC:     permitDESC,
	}
}

func (permitRepo PermitRepo) Get(filter models.PermitFilter, limit, offset int, reversed bool) ([]models.Permit, error) {
	if limit < 0 || offset < 0 {
		return nil, fmt.Errorf("permit_repo.Get: %w: limit or offset cannot be zero", ErrInvalidArg)
	}

	permitWhere := permitRepo.permitSelect

	switch filter {
	case models.ActivePermits:
		permitWhere = permitWhere.Where(permitRepo.activeWhere)
	case models.ExpiredPermits:
		permitWhere = permitWhere.Where(permitRepo.expiredWhere(15))
	case models.ExceptionPermits:
		permitWhere = permitWhere.Where(permitRepo.exceptionWhere)
	case models.AllPermits:
		// do not add a where clause in this case
	}

	orderSQL := permitRepo.permitASC
	if reversed {
		orderSQL = permitRepo.permitDESC
	}

	query, _, err := permitWhere.
		OrderBy(orderSQL).
		Limit(uint64(getBoundedLimit(limit))).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("permit_repo.Get: %w: %v", ErrBuildingQuery, err)
	}

	permits := permitSlice{}
	err = permitRepo.database.driver.Select(&permits, query)
	if err != nil {
		return nil, fmt.Errorf("permit_repo.Get: %w: %v", ErrDatabaseQuery, err)
	}

	return permits.toModels(), nil
}

func (permitRepo PermitRepo) GetAllTotalAmount() (int, error) {
	query, _, err := permitRepo.countSelect.ToSql()
	if err != nil {
		return 0, fmt.Errorf("permit_repo.GetAllTotalAmount: %w: %v", ErrBuildingQuery, err)
	}

	var totalAmount int
	err = permitRepo.database.driver.Get(&totalAmount, query)
	if err != nil {
		return 0, fmt.Errorf("permit_repo.GetAllTotalAmount: %w: %v", ErrDatabaseQuery, err)
	}

	return totalAmount, nil
}

func (permitRepo PermitRepo) GetActiveTotalAmount() (int, error) {
	query, _, err := permitRepo.countSelect.Where(permitRepo.activeWhere).ToSql()
	if err != nil {
		return 0, fmt.Errorf("permit_repo.GetActiveTotalAmount: %w: %v", ErrBuildingQuery, err)
	}

	var totalAmount int
	err = permitRepo.database.driver.Get(&totalAmount, query)
	if err != nil {
		return 0, fmt.Errorf("permit_repo.GetActiveTotalAmount: %w: %v", ErrDatabaseQuery, err)
	}

	return totalAmount, nil
}

func (permitRepo PermitRepo) GetExceptionsTotalAmount() (int, error) {
	query, _, err := permitRepo.countSelect.
		Where(permitRepo.exceptionWhere).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("permit_repo.GetExceptionsTotalAmount: %w: %v", ErrBuildingQuery, err)
	}

	var totalAmount int
	err = permitRepo.database.driver.Get(&totalAmount, query)
	if err != nil {
		return 0, fmt.Errorf("permit_repo.GetExceptionsTotalAmount: %w: %v", ErrDatabaseQuery, err)
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

func (permitRepo PermitRepo) Create(newPermitArgs models.NewPermitArgs) (int, error) {
	const query = `
    INSERT INTO permit(resident_id, car_id, start_ts, end_ts, request_ts, affects_days, exception_reason)
    VALUES($1, $2, $3, $4, extract(epoch from now()), $5, $6)
    RETURNING id
  `

	var id int
	err := permitRepo.database.driver.Get(&id, query, newPermitArgs.ResidentId, newPermitArgs.CarId,
		newPermitArgs.StartDate.Unix(), newPermitArgs.EndDate.Unix(), newPermitArgs.AffectsDays,
		toNullable(newPermitArgs.ExceptionReason))
	if err != nil {
		return 0, fmt.Errorf("permit_repo.Create: %w: %v", ErrDatabaseExec, err)
	}

	return id, nil
}

func (permitRepo PermitRepo) GetActiveOfCarDuring(carId string, startDate, endDate time.Time) ([]models.Permit, error) {
	query, args, err := permitRepo.permitSelect.
		Where("car_id = $1", carId).
		Where("permit.start_ts <= $2", endDate.Unix()).
		Where("permit.end_ts >= $3", startDate.Unix()).
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

func (permitRepo PermitRepo) GetActiveOfResidentDuring(residentId string, startDate, endDate time.Time) ([]models.Permit, error) {
	query, args, err := permitRepo.permitSelect.
		Where("permit.resident_id = $1", residentId).
		Where("permit.start_ts <= $2", endDate.Unix()).
		Where("permit.end_ts >= $3", startDate.Unix()).
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

func (permitRepo PermitRepo) Search(searchStr string, filter models.PermitFilter) ([]models.Permit, error) {
	if searchStr == "" {
		return nil, fmt.Errorf("permit_repo.Search: %w: Empty search argument", ErrInvalidArg)
	}

	permitWhere := permitRepo.permitSelect.
		Where(squirrel.Or{
			squirrel.Expr("CAST(permit.id AS TEXT) ILIKE $1", "%"+searchStr+"%"),
			squirrel.Expr("permit.resident_id ILIKE $1"),
			squirrel.Expr("car.license_plate ILIKE $1"),
			squirrel.Expr("car.color ILIKE $1"),
			squirrel.Expr("car.make ILIKE $1"),
			squirrel.Expr("car.model ILIKE $1"),
		})

	switch filter {
	case models.ActivePermits:
		permitWhere = permitWhere.Where(permitRepo.activeWhere)
	case models.ExpiredPermits:
		permitWhere = permitWhere.Where(permitRepo.expiredWhere(15))
	case models.ExceptionPermits:
		permitWhere = permitWhere.Where(permitRepo.exceptionWhere)
	case models.AllPermits:
		// do not add a where clause in this case
	}

	query, args, err := permitWhere.OrderBy(permitRepo.permitASC).ToSql()
	if err != nil {
		return nil, fmt.Errorf("permit_repo.Search: %w: %v", ErrBuildingQuery, err)
	}

	permits := permitSlice{}
	err = permitRepo.database.driver.Select(&permits, query, args...)
	if err != nil {
		return nil, fmt.Errorf("permit_repo.Search: %w: %v", ErrDatabaseQuery, err)
	}

	return permits.toModels(), nil
}
