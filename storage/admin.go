package storage

import (
	"github.com/dannyvelas/lasvistas_api/models"
)

type admin struct {
	Id           string `db:"id"`
	FirstName    string `db:"first_name"`
	LastName     string `db:"last_name"`
	Email        string `db:"email"`
	Password     string `db:"password"`
	IsPrivileged bool   `db:"is_privileged"`
}

func (admin admin) toModels() models.Admin {
	return models.NewAdmin(
		admin.Id,
		admin.FirstName,
		admin.LastName,
		admin.Email,
		admin.Password,
		admin.IsPrivileged,
	)
}
