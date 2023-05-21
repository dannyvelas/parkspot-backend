package psql

import (
	"database/sql"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage/selectopts"
	"strings"
)

type CarRepo struct {
	database    Database
	carSelect   squirrel.SelectBuilder
	countSelect squirrel.SelectBuilder
}

func NewCarRepo(database Database) CarRepo {
	carSelect := stmtBuilder.Select(
		"car.id",
		"car.resident_id",
		"car.license_plate",
		"car.color",
		"car.make",
		"car.model",
		"car.amt_parking_days_used",
	).From("car")
	countSelect := stmtBuilder.Select("count(*)").From("car")

	return CarRepo{
		database:    database,
		carSelect:   carSelect,
		countSelect: countSelect,
	}
}

func (carRepo CarRepo) GetOne(id string) (models.Car, error) {
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

func (carRepo CarRepo) SelectWhere(carFields models.Car, selectOpts ...selectopts.SelectOpt) ([]models.Car, error) {
	selector := carRepo.carSelect
	for _, opt := range selectOpts {
		selector = opt.Dispatch(carRepo, selector)
	}

	carSelect := selector.Where(rmEmptyVals(squirrel.Eq{
		"resident_id":   carFields.ResidentID,
		"license_plate": carFields.LicensePlate,
		"color":         carFields.Color,
		"make":          carFields.Make,
		"model":         carFields.Model,
	}))
	query, args, err := carSelect.ToSql()
	if err != nil {
		return nil, fmt.Errorf("car_repo.SelectWhere: %w: %v", errs.DBBuildingQuery, err)
	}

	cars := carSlice{}
	err = carRepo.database.driver.Select(&cars, query, args...)
	if err != nil {
		return nil, fmt.Errorf("car_repo.SelectWhere: %w: %v", errs.DBQuery, err)
	}

	return cars.toModels(), nil
}

func (carRepo CarRepo) SelectCountWhere(carFields models.Car, selectOpts ...selectopts.SelectOpt) (int, error) {
	selector := stmtBuilder.Select("count(*)").From("car")
	for _, opt := range selectOpts {
		selector = opt.Dispatch(carRepo, selector)
	}

	countSelect := selector.Where(rmEmptyVals(squirrel.Eq{
		"resident_id":   carFields.ResidentID,
		"license_plate": carFields.LicensePlate,
		"color":         carFields.Color,
		"make":          carFields.Make,
		"model":         carFields.Model,
	}))
	query, args, err := countSelect.ToSql()
	if err != nil {
		return 0, fmt.Errorf("car_repo.SelectCount: %w: %v", errs.DBBuildingQuery, err)
	}

	var totalAmount int
	err = carRepo.database.driver.Get(&totalAmount, query, args...)
	if err != nil {
		return 0, fmt.Errorf("car_repo.GetCount: %w: %v", errs.DBQuery, err)
	}

	return totalAmount, nil
}

func (carRepo CarRepo) Create(desiredCar models.Car) (string, error) {
	updateMap := make(squirrel.Eq)
	if desiredCar.ID != "" {
		updateMap["id"] = desiredCar.ID
	}
	updateMap["resident_id"] = desiredCar.ResidentID
	updateMap["license_plate"] = desiredCar.LicensePlate
	updateMap["color"] = desiredCar.Color
	updateMap["make"] = desiredCar.Make
	updateMap["model"] = desiredCar.Model

	query, args, err := stmtBuilder.Insert("car").SetMap(updateMap).Suffix("RETURNING id").ToSql()
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

func (carRepo CarRepo) Update(carFields models.Car) error {
	carUpdate := stmtBuilder.Update("car").SetMap(rmEmptyVals(squirrel.Eq{
		"license_plate": carFields.LicensePlate,
		"color":         carFields.Color,
		"make":          carFields.Make,
		"model":         carFields.Model,
	}))
	if carFields.AmtParkingDaysUsed != nil {
		carUpdate = carUpdate.Set("amt_parking_days_used", *carFields.AmtParkingDaysUsed)
	}

	query, args, err := carUpdate.Where("car.id = ?", carFields.ID).ToSql()
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
	const query = `DELETE FROM car WHERE id = $1`

	_, err := carRepo.database.driver.Exec(query, id)
	if err != nil {
		return fmt.Errorf("car_repo.Delete: %w: %v", errs.DBExec, err)
	}

	return nil
}

func (carRepo CarRepo) Reset() error {
	_, err := carRepo.database.driver.Exec("DELETE FROM car")
	if err != nil {
		return fmt.Errorf("car_repo.Reset: %w: %v", errs.DBExec, err)
	}

	return nil
}

// implement selectops.Repo
func (carRepo CarRepo) SearchAsSQL(query string) squirrel.Sqlizer {
	lcQuery := strings.ToLower(query)
	return squirrel.Or{
		squirrel.Expr("LOWER(car.resident_id) = ?", lcQuery),
		squirrel.Expr("LOWER(car.license_plate) = ?", lcQuery),
		squirrel.Expr("LOWER(car.color) = ?", lcQuery),
		squirrel.Expr("LOWER(car.make) = ?", lcQuery),
		squirrel.Expr("LOWER(car.model) = ?", lcQuery),
	}
}

func (carRepo CarRepo) StatusAsSQL(status models.Status) (squirrel.Sqlizer, bool) {
	return nil, false
}
