package storage

import (
	"github.com/dannyvelas/parkspot-backend/models"
	"github.com/dannyvelas/parkspot-backend/storage/selectopts"
)

type VisitorRepo interface {
	SelectWhere(models.Visitor, ...selectopts.SelectOpt) ([]models.Visitor, error)
	SelectCountWhere(models.Visitor, ...selectopts.SelectOpt) (int, error)
	Create(desiredVisitor models.Visitor) (string, error)
	Delete(visitorID string) error
	GetOne(visitorID string) (models.Visitor, error)
}
