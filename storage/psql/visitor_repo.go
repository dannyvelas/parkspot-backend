package psql

import (
	"database/sql"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/dannyvelas/lasvistas_api/storage/selectopts"
	"github.com/jmoiron/sqlx"
	"strings"
)

type VisitorRepo struct {
	driver        *sqlx.DB
	visitorSelect squirrel.SelectBuilder
	countSelect   squirrel.SelectBuilder
}

func NewVisitorRepo(driver *sqlx.DB) storage.VisitorRepo {
	visitorSelect := stmtBuilder.Select(
		"id",
		"resident_id",
		"first_name",
		"last_name",
		"relationship",
		"access_start",
		"access_end",
	).From("visitor")
	countSelect := stmtBuilder.Select("count(*)").From("visitor")

	return VisitorRepo{
		driver:        driver,
		visitorSelect: visitorSelect,
		countSelect:   countSelect,
	}
}

func (visitorRepo VisitorRepo) SelectWhere(visitorFields models.Visitor, selectOpts ...selectopts.SelectOpt) ([]models.Visitor, error) {
	selector := visitorRepo.visitorSelect
	for _, opt := range selectOpts {
		selector = opt.Dispatch(visitorRepo, selector)
	}

	visitorSelect := selector.Where(rmEmptyVals(squirrel.Eq{
		"id":          visitorFields.ID,
		"resident_id": visitorFields.ResidentID,
		"first_name":  visitorFields.FirstName,
		"last_name":   visitorFields.LastName,
	}))

	query, args, err := visitorSelect.ToSql()
	if err != nil {
		return nil, fmt.Errorf("visitor_repo.SelectWhere: %w: %v", errs.DBBuildingQuery, err)
	}

	visitors := visitorSlice{}
	err = visitorRepo.driver.Select(&visitors, query, args...)
	if err != nil {
		return nil, fmt.Errorf("visitor_repo.SelectWhere: %w: %v", errs.DBQuery, err)
	}

	return visitors.toModels(), nil
}

func (visitorRepo VisitorRepo) SelectCountWhere(visitorFields models.Visitor, selectOpts ...selectopts.SelectOpt) (int, error) {
	selector := visitorRepo.countSelect
	for _, opt := range selectOpts {
		selector = opt.Dispatch(visitorRepo, selector)
	}

	countSelect := selector.Where(rmEmptyVals(squirrel.Eq{
		"id":          visitorFields.ID,
		"resident_id": visitorFields.ResidentID,
		"first_name":  visitorFields.FirstName,
		"last_name":   visitorFields.LastName,
	}))
	query, args, err := countSelect.ToSql()
	if err != nil {
		return 0, fmt.Errorf("visitor_repo.SelectCountWhere: %w: %v", errs.DBBuildingQuery, err)
	}

	var totalAmount int
	err = visitorRepo.driver.Get(&totalAmount, query, args...)
	if err != nil {
		return 0, fmt.Errorf("visitor_repo.SelectCountWhere: %w: %v", errs.DBQuery, err)
	}

	return totalAmount, nil
}

func (visitorRepo VisitorRepo) Create(desiredVisitor models.Visitor) (string, error) {
	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := sq.
		Insert("visitor").
		SetMap(squirrel.Eq{
			"resident_id":  desiredVisitor.ResidentID,
			"first_name":   desiredVisitor.FirstName,
			"last_name":    desiredVisitor.LastName,
			"relationship": desiredVisitor.Relationship,
			"access_start": desiredVisitor.AccessStart.Unix(),
			"access_end":   desiredVisitor.AccessEnd.Unix(),
		}).
		Suffix("RETURNING visitor.id").
		ToSql()
	if err != nil {
		return "", fmt.Errorf("visitor_repo.Create: %w: %v", errs.DBBuildingQuery, err)
	}

	var visitorID string
	err = visitorRepo.driver.Get(&visitorID, query, args...)
	if err != nil {
		return "", fmt.Errorf("visitor_repo.Create: %w: %v", errs.DBExec, err)
	}

	return visitorID, nil
}

func (visitorRepo VisitorRepo) Delete(visitorID string) error {
	const query = `DELETE FROM visitor WHERE id = $1`

	res, err := visitorRepo.driver.Exec(query, visitorID)
	if err != nil {
		return fmt.Errorf("visitor_repo.Delete: %w: %v", errs.DBExec, err)
	}

	if rowsAffected, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("visitor_repo.Delete: %w: %v", errs.DBGetRowsAffected, err)
	} else if rowsAffected == 0 {
		return fmt.Errorf("visitor_repo.Delete: %w", errs.NewNotFound("visitor"))
	}

	return nil
}

func (visitorRepo VisitorRepo) GetOne(visitorID string) (models.Visitor, error) {
	query, args, err := visitorRepo.visitorSelect.Where("visitor.id = $1", visitorID).ToSql()
	if err != nil {
		return models.Visitor{}, fmt.Errorf("visitor_repo.GetOne: %w: %v", errs.DBBuildingQuery, err)
	}

	visitor := visitor{}
	err = visitorRepo.driver.Get(&visitor, query, args...)
	if err == sql.ErrNoRows {
		return models.Visitor{}, fmt.Errorf("visitor_repo.GetOne: %w", errs.NewNotFound("visitor"))
	} else if err != nil {
		return models.Visitor{}, fmt.Errorf("visitor_repo.GetOne: %w: %v", errs.DBQuery, err)
	}

	return visitor.toModels(), nil
}

// helpers
func (visitorRepo VisitorRepo) SearchAsSQL(query string) squirrel.Sqlizer {
	likeQuery := "%" + strings.ToLower(query) + "%"
	return squirrel.Or{
		squirrel.Expr("visitor.resident_id ILIKE ?", likeQuery),
		squirrel.Expr("visitor.first_name ILIKE ?", likeQuery),
		squirrel.Expr("visitor.last_name ILIKE ?", likeQuery),
	}
}

func (visitorRepo VisitorRepo) StatusAsSQL(status models.Status) (squirrel.Sqlizer, bool) {
	if status == models.ActiveStatus {
		return squirrel.And{
			squirrel.Expr("visitor.access_start <= extract(epoch from now())"),
			squirrel.Expr("visitor.access_end >= extract(epoch from now())"),
		}, true
	}
	return nil, false
}
