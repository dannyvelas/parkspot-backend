package storage

import (
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/util"
	"github.com/google/uuid"
)

type CarRepoMock struct {
	cars []models.Car
}

func NewCarRepoMock() CarRepoMock {
	return CarRepoMock{cars: make([]models.Car, 0)}
}

func (carRepoMock *CarRepoMock) GetOne(id string) (models.Car, error) {
	i := util.Find(carRepoMock.cars, func(car models.Car) bool { return car.ID == id })
	if i < 0 {
		return models.Car{}, errs.NewNotFound("car")
	}
	return carRepoMock.cars[i], nil
}

func (carRepoMock *CarRepoMock) SelectWhere(carFields models.Car) ([]models.Car, error) {
	var carsFound []models.Car
	for _, car := range carRepoMock.cars {
		if (carFields.ID == "" || carFields.ID == car.ID) &&
			(carFields.ResidentID == "" || carFields.ResidentID == car.ResidentID) &&
			(carFields.LicensePlate == "" || carFields.LicensePlate == car.LicensePlate) &&
			(carFields.Color == "" || carFields.Color == car.Color) &&
			(carFields.Make == "" || carFields.Make == car.Make) &&
			(carFields.Model == "" || carFields.Model == car.Model) {
			carsFound = append(carsFound, car)
		}
	}
	return carsFound, nil
}

func (carRepoMock *CarRepoMock) SelectCountWhere(carFields models.Car) (int, error) {
	return len(carRepoMock.cars), nil
}

func (carRepoMock *CarRepoMock) Create(car models.Car) (string, error) {
	if car.ID == "" {
		car.ID = uuid.NewString()
	}
	carRepoMock.cars = append(carRepoMock.cars, car)

	return car.ID, nil
}

func (carRepoMock *CarRepoMock) AddToAmtParkingDaysUsed(id string, days int) error {
	i := util.Find(carRepoMock.cars, func(car models.Car) bool { return car.ID == id })
	if i == -1 {
		return errs.NewNotFound("car")
	}
	car := &carRepoMock.cars[i]

	*car.AmtParkingDaysUsed = *car.AmtParkingDaysUsed + days
	return nil
}

func (carRepoMock *CarRepoMock) Update(carFields models.Car) error {
	i := util.Find(carRepoMock.cars, func(car models.Car) bool { return car.ID == carFields.ID })
	if i < 0 {
		return errs.NewNotFound("car")
	}
	car := &carRepoMock.cars[i]
	if carFields.LicensePlate != "" {
		car.LicensePlate = carFields.LicensePlate
	}
	if carFields.Color != "" {
		car.Color = carFields.Color
	}
	if carFields.Make != "" {
		car.Make = carFields.Make
	}
	if carFields.Model != "" {
		car.Model = carFields.Model
	}
	if carFields.AmtParkingDaysUsed != nil {
		*car.AmtParkingDaysUsed = *carFields.AmtParkingDaysUsed
	}
	return nil
}

func (carRepoMock *CarRepoMock) Delete(id string) error {
	i := util.Find(carRepoMock.cars, func(car models.Car) bool { return car.ID == id })
	if i == -1 {
		return errs.NewNotFound("car")
	}
	cars := carRepoMock.cars

	// replace the element at the index you want to delete with the last element
	cars[i] = cars[len(cars)-1]

	// re-size slice to remove the last element
	carRepoMock.cars = cars[:len(cars)-1]

	return nil
}

func (carRepoMock *CarRepoMock) Reset() error {
	carRepoMock.cars = carRepoMock.cars[:0]
	return nil
}
