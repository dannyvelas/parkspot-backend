package app

import (
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/errs"
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

func (s CarService) Update(updatedFields models.Car) (models.Car, error) {
	if err := updatedFields.ValidateEdit(); err != nil {
		return models.Car{}, err
	}

	// if license plate is being updated to a new one, make sure it's unique
	if updatedFields.LicensePlate != "" {
		if _, err := s.carRepo.GetByLicensePlate(updatedFields.LicensePlate); err == nil {
			return models.Car{}, errs.AlreadyExists("a car with this license plate")
		} else if !errors.Is(err, errs.NotFound) {
			return models.Car{}, fmt.Errorf("car_service.update error getting car by license plate: %v", err)
		}
	}

	err := s.carRepo.Update(updatedFields)
	if err != nil {
		return models.Car{}, fmt.Errorf("error updating car from carRepo: %w", err)
	}

	car, err := s.carRepo.GetOne(updatedFields.ID)
	if err != nil {
		return models.Car{}, fmt.Errorf("error getting car from carRepo: %w", err)
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
