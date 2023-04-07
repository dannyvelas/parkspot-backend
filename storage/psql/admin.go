package psql

import (
	"github.com/dannyvelas/lasvistas_api/models"
)

type admin struct {
	ID           string `db:"id"`
	FirstName    string `db:"first_name"`
	LastName     string `db:"last_name"`
	Email        string `db:"email"`
	Password     string `db:"password"`
	IsPrivileged bool   `db:"is_privileged"`
	TokenVersion int    `db:"token_version"`
}

func (admin admin) toModels() models.Admin {
	return models.NewAdmin(
		admin.ID,
		admin.FirstName,
		admin.LastName,
		admin.Email,
		admin.Password,
		admin.IsPrivileged,
		admin.TokenVersion,
	)
}
