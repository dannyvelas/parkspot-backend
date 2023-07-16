package psql

import (
	"database/sql"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
)

type AdminRepo struct {
	database storage.Database
}

func NewAdminRepo(database storage.Database) AdminRepo {
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
	err := adminRepo.database.Driver().Get(&admin, query, id)
	if err == sql.ErrNoRows {
		return models.Admin{}, fmt.Errorf("admin_repo.GetOne: %w", errs.NewNotFound("admin"))
	} else if err != nil {
		return models.Admin{}, fmt.Errorf("admin_repo.GetOne: %w: %v", errs.DBQueryScanOneRow, err)
	}

	return admin.toModels(), nil
}

func (adminRepo AdminRepo) Update(adminFields models.Admin) error {
	adminUpdate := stmtBuilder.Update("admin").SetMap(rmEmptyVals(squirrel.Eq{
		"first_name":    adminFields.FirstName,
		"last_name":     adminFields.LastName,
		"email":         adminFields.Email,
		"password":      adminFields.Password,
		"is_privileged": adminFields.IsPrivileged,
	}))
	if adminFields.TokenVersion != nil {
		adminUpdate = adminUpdate.Set("token_version", *adminFields.TokenVersion)
	}

	query, args, err := adminUpdate.Where("admin.id = ?", adminFields.ID).ToSql()
	if err != nil {
		return fmt.Errorf("admin_repo.Update: %w: %v", errs.DBBuildingQuery, err)
	}

	_, err = adminRepo.database.Driver().Exec(query, args...)
	if err != nil {
		return fmt.Errorf("admin_repo.Update: %w: %v", errs.DBExec, err)
	}

	return nil
}
