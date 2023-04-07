package storage

import (
	"github.com/dannyvelas/lasvistas_api/models"
)

type VisitorRepo interface {
	Get(onlyActive bool, residentID, search string, limit, offset int) ([]models.Visitor, error)
	GetCount(onlyActive bool, residentID string) (int, error)
	Create(desiredVisitor models.Visitor) (string, error)
	Delete(visitorID string) error
	GetOne(visitorID string) (models.Visitor, error)
}
