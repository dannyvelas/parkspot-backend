package storage

import (
	"github.com/dannyvelas/parkspot-backend/models"
	"github.com/dannyvelas/parkspot-backend/storage/selectopts"
)

type CarRepo interface {
	GetOne(id string) (models.Car, error)
	SelectWhere(carFields models.Car, selectOpts ...selectopts.SelectOpt) ([]models.Car, error)
	SelectCountWhere(carFields models.Car, selectOpts ...selectopts.SelectOpt) (int, error)
	Create(desiredCar models.Car) (string, error)
	AddToAmtParkingDaysUsed(id string, days int) error
	Update(carFields models.Car) error
	Delete(id string) error
	Reset() error // for testing purposes
}
