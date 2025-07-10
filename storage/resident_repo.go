package storage

import (
	"github.com/dannyvelas/parkspot-backend/models"
	"github.com/dannyvelas/parkspot-backend/storage/selectopts"
)

type ResidentRepo interface {
	SelectWhere(residentFields models.Resident, selectOpts ...selectopts.SelectOpt) ([]models.Resident, error)
	SelectCountWhere(residentFields models.Resident, selectOpts ...selectopts.SelectOpt) (int, error)
	AddToAmtParkingDaysUsed(id string, days int) error
	Create(resident models.Resident) error
	Delete(residentID string) error
	Update(residentFields models.Resident) error
	Reset() error // for testing
}
