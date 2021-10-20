package storage

import (
	"database/sql"
	"github.com/dannyvelas/parkspot-api/models"
)

type AdminRepo struct {
	database *Database
}

func NewAdminRepo(database *Database) *AdminRepo {
	return &AdminRepo{database: database}
}

func (adminRepo *AdminRepo) GetOne(id string) (models.Admin, error) {
	const query = `SELECT id, password FROM admins WHERE id = $1`

	var admin models.Admin
	err := adminRepo.database.driver.QueryRow(query, id).
		Scan(&admin.Id, &admin.Password)

	if err != nil && err == sql.ErrNoRows {
		return models.Admin{}, ResourceNotFound
	} else if err != nil {
		return models.Admin{}, err
	}

	return admin, nil
}
