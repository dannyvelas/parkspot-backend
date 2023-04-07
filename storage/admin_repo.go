package storage

import (
	"github.com/dannyvelas/lasvistas_api/models"
)

type AdminRepo interface {
	GetOne(id string) (models.Admin, error)
	Update(adminFields models.Admin) error
}
