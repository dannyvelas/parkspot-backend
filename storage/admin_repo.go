package storage

import (
	"database/sql"
	"fmt"
)

type AdminRepo struct {
	database Database
}

func NewAdminRepo(database Database) AdminRepo {
	return AdminRepo{database: database}
}

func (adminRepo AdminRepo) GetOne(id string) (Admin, error) {
	const query = `SELECT id, password FROM admin WHERE email = $1`

	var admin Admin
	err := adminRepo.database.driver.QueryRow(query, id).
		Scan(&admin.Id, &admin.Password)
	if err == sql.ErrNoRows {
		return Admin{}, fmt.Errorf("admin_repo: GetOne: %w", ErrNoRows)
	} else if err != nil {
		return Admin{}, fmt.Errorf("admin_repo: GetOne: %v", newError(ErrQueryScanOneRow, err))
	}

	return admin, nil
}
