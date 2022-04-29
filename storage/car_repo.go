package storage

import (
	"database/sql"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/lasvistas_api/models"
)

type CarRepo struct {
	database  Database
	carSelect squirrel.SelectBuilder
}

func NewCarRepo(database Database) CarRepo {
	carSelect := squirrel.Select(
		"car.id AS car_id",
		"car.license_plate",
		"car.color",
		"car.make",
		"car.model",
		"car.amt_parking_days_used",
	).From("car")

	return CarRepo{database: database, carSelect: carSelect}
}

func (carRepo CarRepo) GetOne(id string) (models.Car, error) {
	if id == "" {
		return models.Car{}, fmt.Errorf("car_repo.GetOne: %w: Empty ID argument", ErrInvalidArg)
	}

	query, args, err := carRepo.carSelect.Where("car.id = $1", id).ToSql()
	if err != nil {
		return models.Car{}, fmt.Errorf("car_repo.GetOne: %w: %v", ErrBuildingQuery, err)
	}

	car := car{}
	err = carRepo.database.driver.Get(&car, query, args...)
	if err == sql.ErrNoRows {
		return models.Car{}, fmt.Errorf("car_repo.GetOne: %w", ErrNoRows)
	} else if err != nil {
		return models.Car{}, fmt.Errorf("car_repo.GetOne: %w: %v", ErrDatabaseQuery, err)
	}

	return car.toModels(), nil
}

func (carRepo CarRepo) GetByLicensePlate(licensePlate string) (models.Car, error) {
	if licensePlate == "" {
		return models.Car{}, fmt.Errorf("car_repo.GetByLicensePlate: %w: Empty licensePlate argument", ErrInvalidArg)
	}

	query, args, err := carRepo.carSelect.Where("car.license_plate = $1", licensePlate).ToSql()
	if err != nil {
		return models.Car{}, fmt.Errorf("car_repo.GetByLicensePlate: %w: %v", ErrBuildingQuery, err)
	}

	car := car{}
	err = carRepo.database.driver.Get(&car, query, args...)
	if err == sql.ErrNoRows {
		return models.Car{}, fmt.Errorf("car_repo.GetByLicensePlate: %w", ErrNoRows)
	} else if err != nil {
		return models.Car{}, fmt.Errorf("car_repo.GetByLicensePlate: %w: %v", ErrDatabaseQuery, err)
	}

	return car.toModels(), nil
}

func (carRepo CarRepo) Create(createCar models.CreateCar) (string, error) {
	const query = `
    INSERT INTO car(license_plate, color, make, model)
    VALUES($1, $2, $3, $4)
    RETURNING id
  `

	var id string
	err := carRepo.database.driver.Get(&id, query, createCar.LicensePlate, createCar.Color, createCar.Make, createCar.Model)
	if err != nil {
		return "", fmt.Errorf("car_repo.Create: %w: %v", ErrDatabaseExec, err)
	}

	return id, nil
}

func (carRepo CarRepo) AddToAmtParkingDaysUsed(id string, days int) error {
	const query = `
    UPDATE car SET amt_parking_days_used = amt_parking_days_used + $1
    WHERE id = $2
  `

	_, err := carRepo.database.driver.Exec(query, days, id)
	if err != nil {
		return fmt.Errorf("car_repo.AddToAmtParkingDaysUsed: %w: %v", ErrDatabaseExec, err)
	}

	return nil
}
