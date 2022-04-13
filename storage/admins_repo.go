package storage

import (
	"database/sql"
	"fmt"
)

type AdminsRepo struct {
	database Database
}

func NewAdminsRepo(database Database) AdminsRepo {
	return AdminsRepo{database: database}
}

func (adminsRepo AdminsRepo) GetOne(id string) (Admin, error) {
	const query = `SELECT id, password FROM admins WHERE email = $1`

	var admin Admin
	err := adminsRepo.database.driver.QueryRow(query, id).
		Scan(&admin.Id, &admin.Password)
	if err == sql.ErrNoRows {
		return Admin{}, fmt.Errorf("admin_repo: GetOne: %w", ErrNoRows)
	} else if err != nil {
		return Admin{}, fmt.Errorf("admin_repo: GetOne: %v", newError(ErrQueryScanOneRow, err))
	}

	return admin, nil
}
