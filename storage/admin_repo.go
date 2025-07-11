package storage

import (
	"github.com/dannyvelas/parkspot-backend/models"
)

type AdminRepo interface {
	GetOne(id string) (models.Admin, error)
	Update(adminFields models.Admin) error
	Create(desiredAdmin models.Admin) error
	Delete(id string) error
}
