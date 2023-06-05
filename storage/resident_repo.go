package storage

import (
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage/selectopts"
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
