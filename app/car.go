package app

import (
	"errors"
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
	car, err := s.carRepo.GetOne(id)
	if errors.Is(err, storage.ErrNoRows) {
		return models.Car{}, ErrNotFound
	} else if err != nil {
		return models.Car{}, fmt.Errorf("error getting car from car repo: %v", err)
	}

	return car, nil
}

func (s CarService) GetByLicensePlate(licensePlate string) (*models.Car, error) {
	car, err := s.carRepo.GetByLicensePlate(licensePlate)
	if errors.Is(err, storage.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("error getting car from car repo: %v", err)
	}

	return car, nil
}

func (s CarService) Delete(id string) error {
	err := s.carRepo.Delete(id)
	if errors.Is(err, storage.ErrNoRows) {
		return ErrNotFound
	} else if err != nil {
		return fmt.Errorf("error deleting in carRepo: %v", err)
	}

	return nil
}

func (s CarService) Update(id string, editCar models.Car) (models.Car, error) {
	err := s.carRepo.Update(id, editCar)
	if err != nil {
		return models.Car{}, fmt.Errorf("error updating car from carRepo: %v", err)
	}

	car, err := s.carRepo.GetOne(id)
	if err != nil {
		return models.Car{}, fmt.Errorf("error getting car from carRepo: %v", err)
	}

	return car, nil
}

func (s CarService) Upsert(desiredCar models.Car) (models.Car, error) {
	existingCar, err := s.carRepo.GetByLicensePlate(desiredCar.LicensePlate)
	if err != nil && !errors.Is(err, storage.ErrNoRows) { // unexpected error
		return models.Car{}, fmt.Errorf("error getting by licensePlate in carRepo: %v", err)
	} else if errors.Is(err, storage.ErrNoRows) {
		// no-op: if car DNE, this is valid and acceptable
	}

	// car exists and has all fields
	if existingCar != nil && existingCar.Make != "" && existingCar.Model != "" {
		return *existingCar, nil
	}

	// car exists but missing fields
	if existingCar != nil {
		err := s.carRepo.Update(existingCar.ID, desiredCar)
		if err != nil {
			return models.Car{}, fmt.Errorf("error updating car which is missing fields: %v", err)
		}
		newCar := models.NewCar(existingCar.ID, desiredCar.LicensePlate, desiredCar.Color, desiredCar.Make, desiredCar.Model, existingCar.AmtParkingDaysUsed)
		return newCar, nil
	}

	// car DNE
	carID, err := s.carRepo.Create(desiredCar.LicensePlate, desiredCar.Color, desiredCar.Make, desiredCar.Model)
	if err != nil {
		return models.Car{}, fmt.Errorf("error creating car with carRepo: %v", err)
	}

	newCar := models.NewCar(carID, desiredCar.LicensePlate, desiredCar.Color, desiredCar.Make, desiredCar.Model, 0)
	return newCar, nil
}
