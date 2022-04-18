package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/models"
)

type carRepo struct {
	database Database
}

func newCarRepo(database Database) carRepo {
	return carRepo{database: database}
}

func (carRepo carRepo) GetOne(id string) (models.Car, error) {
	if id == "" {
		return models.Car{}, fmt.Errorf("car_repo: GetOne: %w", errMissingFields([]string{"Id"}))
	}

	const query = `
    SELECT
      car.id AS car_id,
      car.license_plate,
      car.color,
      car.make,
      car.model
    FROM car
    WHERE car.id = $1
  `

	car := car{}
	err := carRepo.database.driver.Get(&car, query, id)
	if err == sql.ErrNoRows {
		return models.Car{}, fmt.Errorf("car_repo: GetOne: %w", ErrNoRows)
	} else if err != nil {
		return models.Car{}, fmt.Errorf("car_repo: GetOne: %w", newError(ErrDatabaseQuery, err))
	}

	return car.toModels(), nil
}

func (carRepo carRepo) CreateIfNotExists(inCar models.Car) (models.Car, error) {
	// not checking for empty fields because that already happens in GetOne and Create
	outCar, err := carRepo.GetOne(inCar.Id)
	if err != nil && !errors.Is(err, ErrNoRows) {
		return models.Car{}, fmt.Errorf("car_repo: CreateIfNotExists: %w", err)
	} else if errors.Is(err, ErrNoRows) {
		outCar, err = carRepo.Create(inCar)
		if err != nil {
			return models.Car{}, fmt.Errorf("car_repo: CreateIfNotExists: %w", err)
		}
	}

	return outCar, nil
}

func (carRepo carRepo) Create(car models.Car) (models.Car, error) {
	emptyFields := car.EmptyFields()
	if len(emptyFields) > 0 {
		return models.Car{}, fmt.Errorf("car_repo: Create: %w: %v", ErrMissingField, emptyFields)
	}

	if err := car.Validate(); err != nil {
		return models.Car{}, fmt.Errorf("car_repo: Create: %w: %v", ErrInvalidField, err)
	}

	const query = `
    INSERT INTO car(id, license_plate, color, make, model)
    VALUES($1, $2, $3, $4, $5);
  `

	_, err := carRepo.database.driver.Exec(query, car.Id, car.LicensePlate, car.Color,
		car.Make, car.Model)
	if err != nil {
		return models.Car{}, fmt.Errorf("car_repo: Create: %w", newError(ErrDatabaseExec, err))
	}

	return car, nil
}
