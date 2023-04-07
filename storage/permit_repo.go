package storage

import (
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage/selectopts"
)

type PermitRepo interface {
	SelectWhere(permitFields models.Permit, selectOpts ...selectopts.SelectOpt) ([]models.Permit, error)
	SelectCountWhere(permitFields models.Permit, selectOpts ...selectopts.SelectOpt) (int, error)
	GetOne(id int) (models.Permit, error)
	Create(desiredPermit models.Permit) (int, error)
	Delete(id int) error
	Update(permitFields models.Permit) error
}
