package storage

import (
	"database/sql"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/lasvistas_api/errs"
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
		return models.Car{}, fmt.Errorf("car_repo.GetOne: %w: Empty ID argument", errs.DBInvalidArg)
	}

	query, args, err := carRepo.carSelect.Where("car.id = $1", id).ToSql()
	if err != nil {
		return models.Car{}, fmt.Errorf("car_repo.GetOne: %w: %v", errs.DBBuildingQuery, err)
	}

	car := car{}
	err = carRepo.database.driver.Get(&car, query, args...)
	if err == sql.ErrNoRows {
		return models.Car{}, fmt.Errorf("car_repo.GetOne: %w", errs.NewNotFound("car"))
	} else if err != nil {
		return models.Car{}, fmt.Errorf("car_repo.GetOne: %w: %v", errs.DBQuery, err)
	}

	return car.toModels(), nil
}

func (carRepo CarRepo) GetByLicensePlate(licensePlate string) (*models.Car, error) {
	if licensePlate == "" {
		return nil, fmt.Errorf("car_repo.GetByLicensePlate: %w: Empty licensePlate argument", errs.DBInvalidArg)
	}

	query, args, err := carRepo.carSelect.Where("car.license_plate = $1", licensePlate).ToSql()
	if err != nil {
		return nil, fmt.Errorf("car_repo.GetByLicensePlate: %w: %v", errs.DBBuildingQuery, err)
	}

	car := car{}
	err = carRepo.database.driver.Get(&car, query, args...)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("car_repo.GetByLicensePlate: %w", errs.NewNotFound("car"))
	} else if err != nil {
		return nil, fmt.Errorf("car_repo.GetByLicensePlate: %w: %v", errs.DBQuery, err)
	}

	asModels := car.toModels()

	return &asModels, nil
}

func (carRepo CarRepo) Create(desiredCar models.Car) (string, error) {
	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	updateMap := make(squirrel.Eq)
	if desiredCar.ID != "" {
		updateMap["id"] = desiredCar.ID
	}
	updateMap["resident_id"] = desiredCar.ResidentID
	updateMap["license_plate"] = desiredCar.LicensePlate
	updateMap["color"] = desiredCar.Color
	updateMap["make"] = desiredCar.Make
	updateMap["model"] = desiredCar.Model

	query, args, err := sq.
		Insert("car").
		SetMap(updateMap).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return "", fmt.Errorf("car_repo.Create: %w: %v", errs.DBBuildingQuery, err)
	}

	var id string
	err = carRepo.database.driver.Get(&id, query, args...)
	if err != nil {
		return "", fmt.Errorf("car_repo.Create: %w: %v", errs.DBExec, err)
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
		return fmt.Errorf("car_repo.AddToAmtParkingDaysUsed: %w: %v", errs.DBExec, err)
	}

	return nil
}

func (carRepo CarRepo) Update(editCar models.Car) error {
	squirrel := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	carUpdate := squirrel.Update("car")

	if editCar.LicensePlate != "" {
		carUpdate = carUpdate.Set("license_plate", editCar.LicensePlate)
	}
	if editCar.Color != "" {
		carUpdate = carUpdate.Set("color", editCar.Color)
	}
	if editCar.Make != "" {
		carUpdate = carUpdate.Set("make", editCar.Make)
	}
	if editCar.Model != "" {
		carUpdate = carUpdate.Set("model", editCar.Model)
	}

	query, args, err := carUpdate.Where("car.id = ?", editCar.ID).ToSql()
	if err != nil {
		return fmt.Errorf("car_repo.Update: %w: %v", errs.DBBuildingQuery, err)
	}

	_, err = carRepo.database.driver.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("car_repo.Update: %w: %v", errs.DBExec, err)
	}

	return nil
}

func (carRepo CarRepo) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("car_repo.Delete: %w: empty string ID argument", errs.DBInvalidArg)
	}
	const query = `DELETE FROM car WHERE id = $1`

	_, err := carRepo.database.driver.Exec(query, id)
	if err != nil {
		return fmt.Errorf("car_repo.Delete: %w: %v", errs.DBExec, err)
	}

	return nil
}

func (carRepo CarRepo) GetByResidentID(residentID string) ([]models.Car, error) {
	query, args, err := carRepo.carSelect.Where("car.residentID = $1", residentID).ToSql()
	if err != nil {
		return nil, fmt.Errorf("car_repo.GetByLicensePlate: %w: %v", errs.DBBuildingQuery, err)
	}

	cars := carSlice{}
	err = carRepo.database.driver.Select(&cars, query, args...)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("car_repo.GetByLicensePlate: %w", errs.NewNotFound("car"))
	} else if err != nil {
		return nil, fmt.Errorf("car_repo.GetByLicensePlate: %w: %v", errs.DBQuery, err)
	}

	return cars.toModels(), nil
}
