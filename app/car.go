package app

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/models/validator"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/dannyvelas/lasvistas_api/storage/selectopts"
	"github.com/dannyvelas/lasvistas_api/util"
)

type CarService struct {
	carRepo storage.CarRepo
}

func NewCarService(carRepo storage.CarRepo) CarService {
	return CarService{
		carRepo: carRepo,
	}
}

func (s CarService) GetAll(limit, page int, reversed bool, search, residentID string) (models.ListWithMetadata[models.Car], error) {
	boundedLimit, offset := getBoundedLimitAndOffset(limit, page)

	allCars, err := s.carRepo.SelectWhere(models.Car{ResidentID: residentID},
		selectopts.WithLimitAndOffset(boundedLimit, offset),
		selectopts.WithReversed(reversed),
		selectopts.WithSearch(search),
	)
	if err != nil {
		return models.ListWithMetadata[models.Car]{}, fmt.Errorf("error getting cars from car repo: %v", err)
	}

	totalAmount, err := s.carRepo.SelectCountWhere(models.Car{ResidentID: residentID},
		selectopts.WithSearch(search),
	)
	if err != nil {
		return models.ListWithMetadata[models.Car]{}, fmt.Errorf("error getting total amount from car repo: %v", err)
	}

	return models.NewListWithMetadata(allCars, totalAmount), nil
}

func (s CarService) GetOne(id string) (models.Car, error) {
	if id == "" {
		return models.Car{}, errs.MissingIDField
	}
	return s.carRepo.GetOne(id)
}

func (s CarService) Delete(id string) error {
	if id == "" {
		return errs.MissingIDField
	}
	return s.carRepo.Delete(id)
}

func (s CarService) Update(updatedFields models.Car) (models.Car, error) {
	if updatedFields.ID == "" {
		return models.Car{}, errs.MissingIDField
	}
	if !util.IsUUIDV4(updatedFields.ID) {
		return models.Car{}, errs.IDNotUUID
	}
	if updatedFields.LicensePlate == "" && updatedFields.Color == "" && updatedFields.Make == "" && updatedFields.Model == "" && updatedFields.AmtParkingDaysUsed == nil {
		return models.Car{}, errs.AllEditFieldsEmpty("licensePlate, color, make, model, amtParkingDaysUsed")
	}
	if err := validator.EditCar.Run(updatedFields); err != nil {
		return models.Car{}, err
	}

	// if license plate is being updated to a new one, make sure it's unique
	if updatedFields.LicensePlate != "" {
		if cars, err := s.carRepo.SelectWhere(models.Car{LicensePlate: updatedFields.LicensePlate}); err != nil {
			return models.Car{}, fmt.Errorf("car_service.update error getting car by license plate: %v", err)
		} else if len(cars) != 0 && cars[0].ID != updatedFields.ID {
			// it's possible that the car we found with the same licensePlate in the database is the car we're currently updating
			// in this case, don't raise an error
			return models.Car{}, errs.NewAlreadyExists("a car with this licensePlate: " + updatedFields.LicensePlate)
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
	if err := validator.CreateCar.Run(desiredCar); err != nil {
		return models.Car{}, err
	}
	if desiredCar.ID != "" {
		if cars, err := s.carRepo.SelectWhere(models.Car{ID: desiredCar.ID}); err != nil {
			return models.Car{}, fmt.Errorf("carService.Create: error getting car by id: %v", err)
		} else if len(cars) != 0 {
			return models.Car{}, errs.NewAlreadyExists("a car with this ID: " + desiredCar.ID)
		}
	}

	if cars, err := s.carRepo.SelectWhere(models.Car{LicensePlate: desiredCar.LicensePlate}); err != nil {
		return models.Car{}, fmt.Errorf("carService.createCar error getting car by licensePlate: %v", err)
	} else if len(cars) != 0 {
		return models.Car{}, errs.NewAlreadyExists("a car with this licensePlate: " + desiredCar.LicensePlate)
	}

	carID, err := s.carRepo.Create(desiredCar)
	if err != nil {
		return models.Car{}, fmt.Errorf("error creating car with carRepo: %v", err)
	}

	newCar := models.NewCar(carID, desiredCar.ResidentID, desiredCar.LicensePlate, desiredCar.Color, desiredCar.Make, desiredCar.Model, 0)
	return newCar, nil
}
