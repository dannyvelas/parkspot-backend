package psql

import (
	"database/sql"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/parkspot-backend/errs"
	"github.com/dannyvelas/parkspot-backend/models"
	"github.com/dannyvelas/parkspot-backend/storage"
	"github.com/jmoiron/sqlx"
)

type AdminRepo struct {
	driver *sqlx.DB
}

func NewAdminRepo(driver *sqlx.DB) storage.AdminRepo {
	return AdminRepo{driver}
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
	err := adminRepo.driver.Get(&admin, query, id)
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

	_, err = adminRepo.driver.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("admin_repo.Update: %w: %v", errs.DBExec, err)
	}

	return nil
}

func (adminRepo AdminRepo) Create(desiredAdmin models.Admin) error {
	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := sq.
		Insert("admin").
		SetMap(squirrel.Eq{
			"id":            desiredAdmin.ID,
			"first_name":    desiredAdmin.FirstName,
			"last_name":     desiredAdmin.LastName,
			"email":         desiredAdmin.Email,
			"password":      desiredAdmin.Password,
			"is_privileged": desiredAdmin.IsPrivileged,
		}).ToSql()
	if err != nil {
		return fmt.Errorf("admin_repo.Create: %w: %v", errs.DBBuildingQuery, err)
	}

	_, err = adminRepo.driver.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("admin_repo.Create: %w: %v", errs.DBExec, err)
	}

	return nil
}

func (adminRepo AdminRepo) Delete(id string) error {
	const query = `DELETE FROM admin WHERE id = $1`

	res, err := adminRepo.driver.Exec(query, id)
	if err != nil {
		return fmt.Errorf("admin_repo.Delete: %w: %v", errs.DBExec, err)
	}

	if rowsAffected, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("admin_repo.Delete: %w: %v", errs.DBGetRowsAffected, err)
	} else if rowsAffected == 0 {
		return fmt.Errorf("admin_repo.Delete: %w", errs.NewNotFound("admin"))
	}

	return nil
}
