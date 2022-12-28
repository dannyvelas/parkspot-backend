package storage

import (
	"database/sql"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/errs"
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
		return models.Admin{}, fmt.Errorf("admin_repo.GetOne: %w: Empty ID", errs.DBInvalidArg)
	}

	const query = `
    SELECT
      id, first_name, last_name, email, password, is_privileged, token_version
    FROM admin
    WHERE LOWER(id) = LOWER($1)
  `

	var admin admin
	err := adminRepo.database.driver.Get(&admin, query, id)
	if err == sql.ErrNoRows {
		return models.Admin{}, fmt.Errorf("admin_repo.GetOne: %w", errs.NotFound("admin"))
	} else if err != nil {
		return models.Admin{}, fmt.Errorf("admin_repo.GetOne: %w: %v", errs.DBQueryScanOneRow, err)
	}

	return admin.toModels(), nil
}

func (adminRepo AdminRepo) SetPassword(id string, password string) error {
	if id == "" {
		return fmt.Errorf("admin_repo.GetOne: %w: Empty ID argument", errs.DBInvalidArg)
	} else if password == "" {
		return fmt.Errorf("admin_repo.GetOne: %w: Emtpy Password argument", errs.DBInvalidArg)
	}

	const query = `UPDATE admin SET password = $1 WHERE id = $2`
	_, err := adminRepo.database.driver.Exec(query, password, id)
	if err != nil {
		return fmt.Errorf("admin_repo.SetPasswordFor: %w: %v", errs.DBExec, err)
	}

	return nil
}
