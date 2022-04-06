package storage

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/models"
)

type PermitsRepo struct {
	database Database
}

func NewPermitsRepo(database Database) PermitsRepo {
	return PermitsRepo{database: database}
}

func (permitsRepo PermitsRepo) GetActive(limit, offset uint) ([]models.Permit, error) {
	const query = `
    SELECT
      permits.id,
      permits.resident_id,
      permits.license_plate,
      cars.color_and_model,
      permits.start_date,
      permits.end_date,
      permits.request_date,
      permits.affects_days
    FROM permits
    LEFT JOIN cars ON
      permits.license_plate = cars.license_plate 
    WHERE
      permits.start_date <= EXTRACT(epoch FROM NOW())
      AND permits.end_date >= EXTRACT(epoch FROM NOW())
    LIMIT $1
    OFFSET $2
  `

	boundedLimit := getBoundedLimit(limit)
	rows, err := permitsRepo.database.driver.Query(query, boundedLimit, offset)
	if err != nil {
		return nil, fmt.Errorf("permits_repo: GetActive: %v", newError(ErrDatabaseQuery, err))
	}
	defer rows.Close()

	permits := []models.Permit{}
	for rows.Next() {
		var permit models.Permit

		err := rows.Scan(
			&permit.Id,
			&permit.ResidentId,
			&permit.LicensePlate,
			&permit.ColorAndModel,
			&permit.StartDate,
			&permit.EndDate,
			&permit.RequestDate,
			&permit.AffectsDays,
		)
		if err != nil {
			return nil, fmt.Errorf("permits_repo: GetActive: %v", newError(ErrScanningRow, err))
		}

		permits = append(permits, permit)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("permits_repo: GetActive: %v", newError(ErrIterating, err))
	}

	return permits, nil
}

func (permitsRepo PermitsRepo) GetAll(limit, offset uint) ([]models.Permit, error) {
	const query = `
    SELECT
      permits.id,
      permits.resident_id,
      permits.license_plate,
      cars.color_and_model,
      permits.start_date,
      permits.end_date,
      permits.request_date,
      permits.affects_days
    FROM permits
    LEFT JOIN cars ON
      permits.license_plate = cars.license_plate 
    LIMIT $1
    OFFSET $2
  `

	boundedLimit := getBoundedLimit(limit)
	rows, err := permitsRepo.database.driver.Query(query, boundedLimit, offset)
	if err != nil {
		return nil, fmt.Errorf("permits_repo: GetAll: %v", newError(ErrDatabaseQuery, err))
	}
	defer rows.Close()

	permits := []models.Permit{}
	for rows.Next() {
		var permit models.Permit
		err := rows.Scan(
			&permit.Id,
			&permit.ResidentId,
			&permit.LicensePlate,
			&permit.ColorAndModel,
			&permit.StartDate,
			&permit.EndDate,
			&permit.RequestDate,
			&permit.AffectsDays,
		)

		if err != nil {
			return nil, fmt.Errorf("permits_repo: GetAll: %v", newError(ErrScanningRow, err))
		}

		permits = append(permits, permit)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("permits_repo: GetAll: %v", newError(ErrIterating, err))
	}

	return permits, nil
}

func (permitsRepo PermitsRepo) deleteAll() (int64, error) {
	query := "DELETE FROM permits"
	res, err := permitsRepo.database.driver.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("permits_repo: deleteAll: %v", newError(ErrDatabaseExec, err))
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("permits_repo: deleteAll: %v", newError(ErrGetRowsAffected, err))
	}

	return rowsAffected, nil
}
