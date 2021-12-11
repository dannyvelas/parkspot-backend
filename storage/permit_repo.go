package storage

import (
	"github.com/dannyvelas/parkspot-api/models"
)

type PermitRepo struct {
	database *Database
}

func NewPermitRepo(database *Database) *PermitRepo {
	return &PermitRepo{database: database}
}

func (permitRepo *PermitRepo) GetActive(limit, offset int) ([]models.Permit, error) {
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

	rows, err := permitRepo.database.driver.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permits []models.Permit
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
			return nil, err
		}

		permits = append(permits, permit)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permits, nil
}

func (permitRepo *PermitRepo) GetAll(limit, offset int) ([]models.Permit, error) {
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

	rows, err := permitRepo.database.driver.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permits []models.Permit
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
			return nil, err
		}

		permits = append(permits, permit)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permits, nil
}
