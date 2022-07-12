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
	if id == "" {
		return models.Admin{}, fmt.Errorf("admin_repo.GetOne: %w: Empty ID argument", ErrInvalidArg)
	}

	const query = `SELECT id, first_name, last_name, email, password, is_privileged FROM admin WHERE id = $1`

	var admin admin
	err := adminRepo.database.driver.Get(&admin, query, id)
	if err == sql.ErrNoRows {
		return models.Admin{}, fmt.Errorf("admin_repo.GetOne: %w", ErrNoRows)
	} else if err != nil {
		return models.Admin{}, fmt.Errorf("admin_repo.GetOne: %w: %v", ErrQueryScanOneRow, err)
	}

	return admin.toModels(), nil
}

func (adminRepo AdminRepo) SetPasswordFor(id string, password string) error {
	if id == "" {
		return fmt.Errorf("admin_repo.GetOne: %w: Empty ID argument", ErrInvalidArg)
	} else if password == "" {
		return fmt.Errorf("admin_repo.GetOne: %w: Emtpy Password argument", ErrInvalidArg)
	}

	const query = `UPDATE admin SET password = $1 WHERE id = $2`
	_, err := adminRepo.database.driver.Exec(query, password, id)
	if err != nil {
		return fmt.Errorf("admin_repo.SetPasswordFor: %w: %v", ErrDatabaseExec, err)
	}

	return nil
}
