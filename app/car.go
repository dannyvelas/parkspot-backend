package app

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
)

type CarService struct {
	carRepo storage.CarRepo
}

func NewCarService(carRepo storage.CarRepo) CarService {
	return CarService{
		carRepo: carRepo,
	}
}

func (s CarService) upsertCar(existingCar *models.Car, desiredCar models.CreateCar) (models.Car, error) {
	// car exits but missing fields
	if existingCar != nil {
		err := s.carRepo.Update(existingCar.Id, desiredCar.Color, desiredCar.Make, desiredCar.Model)
		if err != nil {
			return models.Car{}, fmt.Errorf("error updating car which is missing fields: %v", err)
		}
		newCar := models.NewCar(existingCar.Id, desiredCar.LicensePlate, desiredCar.Color, desiredCar.Make, desiredCar.Model, existingCar.AmtParkingDaysUsed)
		return newCar, nil
	}

	// car DNE
	carId, err := s.carRepo.Create(desiredCar.LicensePlate, desiredCar.Color, desiredCar.Make, desiredCar.Model)
	if err != nil {
		return models.Car{}, fmt.Errorf("error creating car with carRepo: %v", err)
	}

	newCar := models.NewCar(carId, desiredCar.LicensePlate, desiredCar.Color, desiredCar.Make, desiredCar.Model, 0)
	return newCar, nil
}
