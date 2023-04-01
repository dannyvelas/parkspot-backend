package app

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
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

func (s CarService) GetOne(id string) (models.Car, error) {
	return s.carRepo.GetOne(id)
}

func (s CarService) Delete(id string) error {
	return s.carRepo.Delete(id)
}

func (s CarService) Update(updatedFields models.Car) (models.Car, error) {
	if updatedFields.ID == "" {
		return models.Car{}, errs.MissingIDField
	}
	if !util.IsUUIDV4(updatedFields.ID) {
		return models.Car{}, errs.IDNotUUID
	}
	if updatedFields.LicensePlate == "" && updatedFields.Color == "" && updatedFields.Make == "" && updatedFields.Model == "" {
		return models.Car{}, errs.AllEditFieldsEmpty("licensePlate, color, make, model")
	}
	if err := models.EditCarValidator.Run(updatedFields); err != nil {
		return models.Car{}, err
	}

	// if license plate is being updated to a new one, make sure it's unique
	if updatedFields.LicensePlate != "" {
		if cars, err := s.carRepo.SelectWhere(models.Car{LicensePlate: updatedFields.LicensePlate}); err != nil {
			return models.Car{}, fmt.Errorf("car_service.update error getting car by license plate: %v", err)
		} else if len(cars) != 0 {
			return models.Car{}, errs.AlreadyExists("a car with this license plate")
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
	if err := models.CreateCarValidator.Run(desiredCar); err != nil {
		return models.Car{}, err
	}

	carID, err := s.carRepo.Create(desiredCar)
	if err != nil {
		return models.Car{}, fmt.Errorf("error creating car with carRepo: %v", err)
	}

	newCar := models.NewCar(carID, desiredCar.ResidentID, desiredCar.LicensePlate, desiredCar.Color, desiredCar.Make, desiredCar.Model, 0)
	return newCar, nil
}

func (s CarService) GetOfResident(residentID string) (models.ListWithMetadata[models.Car], error) {
	if err := models.IsResidentID(residentID); err != nil {
		return models.ListWithMetadata[models.Car]{}, err
	}

	cars, err := s.carRepo.SelectWhere(models.Car{ResidentID: residentID})
	if err != nil {
		return models.ListWithMetadata[models.Car]{}, fmt.Errorf("Error querying carRepo when getting cars of resident: %v", err)
	}

	count, err := s.carRepo.SelectCountWhere(models.Car{ResidentID: residentID})
	if err != nil {
		return models.ListWithMetadata[models.Car]{}, fmt.Errorf("Error querying carRepo when getting amount of cars of resident: %v", err)
	}

	return models.NewListWithMetadata(cars, count), nil
}
