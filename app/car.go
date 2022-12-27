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
	car, err := s.carRepo.GetOne(id)
	if errors.Is(err, storage.ErrNoRows) {
		return models.Car{}, errs.NotFound
	} else if err != nil {
		return models.Car{}, fmt.Errorf("error getting car from car repo: %v", err)
	}

	return car, nil
}

func (s CarService) GetByLicensePlate(licensePlate string) (*models.Car, error) {
	car, err := s.carRepo.GetByLicensePlate(licensePlate)
	if errors.Is(err, storage.ErrNoRows) {
		return nil, errs.NotFound
	} else if err != nil {
		return nil, fmt.Errorf("error getting car from car repo: %v", err)
	}

	return car, nil
}

func (s CarService) Delete(id string) error {
	err := s.carRepo.Delete(id)
	if errors.Is(err, storage.ErrNoRows) {
		return errs.NotFound
	} else if err != nil {
		return fmt.Errorf("error deleting in carRepo: %v", err)
	}

	return nil
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
	carID, err := s.carRepo.Create(desiredCar)
	if err != nil {
		return models.Car{}, fmt.Errorf("error creating car with carRepo: %v", err)
	}

	newCar := models.NewCar(carID, desiredCar.LicensePlate, desiredCar.Color, desiredCar.Make, desiredCar.Model, 0)
	return newCar, nil
}
