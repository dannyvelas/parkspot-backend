package storage

import (
	"database/sql"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/models"
)

type CarRepo struct {
	database Database
}

func NewCarRepo(database Database) CarRepo {
	return CarRepo{database: database}
}

func (carRepo CarRepo) GetOne(id string) (models.Car, error) {
	if id == "" {
		return models.Car{}, fmt.Errorf("car_repo.GetOne: %w: Empty ID argument", ErrInvalidArg)
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
		return models.Car{}, fmt.Errorf("car_repo.GetOne: %w", ErrNoRows)
	} else if err != nil {
		return models.Car{}, fmt.Errorf("car_repo.GetOne: %w: %v", ErrDatabaseQuery, err)
	}

	return car.toModels(), nil
}

func (carRepo CarRepo) Create(createCar models.CreateCar) (models.Car, error) {
	const query = `
    INSERT INTO car(license_plate, color, make, model)
    VALUES($1, $2, $3, $4);
    RETURNING id
  `

	var id string
	err := carRepo.database.driver.Get(&id, query, createCar.LicensePlate, createCar.Color, createCar.Make, createCar.Model)
	if err != nil {
		return models.Car{}, fmt.Errorf("car_repo.Create: %w: %v", ErrDatabaseExec, err)
	}

	return createCar.ToCar(id), nil
}
