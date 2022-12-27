package app

import (
	"errors"
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

func (s CarService) GetOne(id string) (models.Car, *errs.ApiErr) {
	car, err := s.carRepo.GetOne(id)
	if errors.Is(err, storage.ErrNoRows) {
		return models.Car{}, errs.NotFound("car")
	} else if err != nil {
		return models.Car{}, errs.Internalf("error getting car from car repo: %v", err)
	}

	return car, nil
}

func (s CarService) GetByLicensePlate(licensePlate string) (*models.Car, *errs.ApiErr) {
	car, err := s.carRepo.GetByLicensePlate(licensePlate)
	if errors.Is(err, storage.ErrNoRows) {
		return nil, errs.NotFound("car")
	} else if err != nil {
		return nil, errs.Internalf("error getting car from car repo: %v", err)
	}

	return car, nil
}

func (s CarService) Delete(id string) error {
	err := s.carRepo.Delete(id)
	if errors.Is(err, storage.ErrNoRows) {
		return errs.NotFound("car")
	} else if err != nil {
		return errs.Internalf("error deleting in carRepo: %v", err)
	}

	return nil
}

func (s CarService) Update(id string, updatedFields models.Car) (models.Car, *errs.ApiErr) {
	if err := updatedFields.ValidateEdit(); err != nil {
		return models.Car{}, err
	}

	err := s.carRepo.Update(id, updatedFields)
	if err != nil {
		return models.Car{}, errs.Internalf("error updating car from carRepo: %v", err)
	}

	car, err := s.carRepo.GetOne(id)
	if err != nil {
		return models.Car{}, errs.Internalf("error getting car from carRepo: %v", err)
	}

	return car, nil
}

func (s CarService) Create(desiredCar models.Car) (models.Car, *errs.ApiErr) {
	carID, err := s.carRepo.Create(desiredCar)
	if err != nil {
		return models.Car{}, errs.Internalf("error creating car with carRepo: %v", err)
	}

	newCar := models.NewCar(carID, desiredCar.LicensePlate, desiredCar.Color, desiredCar.Make, desiredCar.Model, 0)
	return newCar, nil
}
