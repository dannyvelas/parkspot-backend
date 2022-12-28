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

func (s CarService) GetOne(id string) (models.Car, error) {
	return s.carRepo.GetOne(id)
}

func (s CarService) GetByLicensePlate(licensePlate string) (*models.Car, error) {
	return s.carRepo.GetByLicensePlate(licensePlate)
}

func (s CarService) Delete(id string) error {
	return s.carRepo.Delete(id)
}

func (s CarService) Update(id string, updatedFields models.Car) (models.Car, error) {
	if err := updatedFields.ValidateEdit(); err != nil {
		return models.Car{}, err
	}

	err := s.carRepo.Update(id, updatedFields)
	if err != nil {
		return models.Car{}, fmt.Errorf("error updating car from carRepo: %v", err)
	}

	car, err := s.carRepo.GetOne(id)
	if err != nil {
		return models.Car{}, fmt.Errorf("error getting car from carRepo: %v", err)
	}

	return car, nil
}

func (s CarService) Create(desiredCar models.Car) (models.Car, error) {
	if err := desiredCar.ValidateCreation(); err != nil {
		return models.Car{}, err
	}

	carID, err := s.carRepo.Create(desiredCar)
	if err != nil {
		return models.Car{}, fmt.Errorf("error creating car with carRepo: %v", err)
	}

	newCar := models.NewCar(carID, desiredCar.ResidentID, desiredCar.LicensePlate, desiredCar.Color, desiredCar.Make, desiredCar.Model, 0)
	return newCar, nil
}
