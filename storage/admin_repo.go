package storage

import (
	"database/sql"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/models"
)

type AdminRepo struct {
	database Database
}

func NewAdminRepo(database Database) AdminRepo {
	return AdminRepo{database: database}
}

func (adminRepo AdminRepo) GetOne(id string) (models.Admin, error) {
	const query = `SELECT id, first_name, last_name, email, password, is_privileged FROM admin WHERE email = $1`

	var admin admin
	err := adminRepo.database.driver.Get(&admin, query, id)
	if err == sql.ErrNoRows {
		return models.Admin{}, fmt.Errorf("admin_repo.GetOne: %w", ErrNoRows)
	} else if err != nil {
		return models.Admin{}, fmt.Errorf("admin_repo.GetOne: %w: %v", ErrQueryScanOneRow, err)
	}

	return admin.toModels(), nil
}
