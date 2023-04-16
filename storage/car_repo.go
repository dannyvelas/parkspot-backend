package storage

import (
	"github.com/dannyvelas/lasvistas_api/models"
)

type CarRepo interface {
	GetOne(id string) (models.Car, error)
	SelectWhere(carFields models.Car) ([]models.Car, error)
	SelectCountWhere(carFields models.Car) (int, error)
	Create(desiredCar models.Car) (string, error)
	AddToAmtParkingDaysUsed(id string, days int) error
	Update(carFields models.Car) error
	Delete(id string) error
	Reset() // for testing purposes
}
