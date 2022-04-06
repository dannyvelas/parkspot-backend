package storage

import (
	"database/sql"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/models"
)

type AdminsRepo struct {
	database Database
}

func NewAdminsRepo(database Database) AdminsRepo {
	return AdminsRepo{database: database}
}

func (adminsRepo AdminsRepo) GetOne(id string) (models.Admin, error) {
	const query = `SELECT id, password FROM admins WHERE email = $1`

	var admin models.Admin
	err := adminsRepo.database.driver.QueryRow(query, id).
		Scan(&admin.Id, &admin.Password)
	if err == sql.ErrNoRows {
		return models.Admin{}, fmt.Errorf("admin_repo: GetOne: %w", ErrNoRows)
	} else if err != nil {
		return models.Admin{}, fmt.Errorf("admin_repo: GetOne: %v", newError(ErrQueryScanOneRow, err))
	}

	return admin, nil
}
