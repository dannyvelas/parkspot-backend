package storage

import (
	"database/sql"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/lasvistas_api/models"
)

type VisitorRepo struct {
	database      Database
	visitorSelect squirrel.SelectBuilder
	whereActive   squirrel.Sqlizer
}

func NewVisitorRepo(database Database) VisitorRepo {
	visitorSelect := squirrel.Select(
		"id",
		"resident_id",
		"first_name",
		"last_name",
		"relationship",
		"access_start",
		"access_end",
	).From("visitor")

	whereActive := squirrel.And{
		squirrel.Expr("visitor.access_start <= extract(epoch from now())"),
		squirrel.Expr("visitor.access_end >= extract(epoch from now())"),
	}

	return VisitorRepo{
		database:      database,
		visitorSelect: visitorSelect,
		whereActive:   whereActive,
	}
}

func (visitorRepo VisitorRepo) Get(onlyActive bool, residentID, search string, limit, offset int) ([]models.Visitor, error) {
	if limit < 0 || offset < 0 {
		return nil, fmt.Errorf("visitor_repo.Get: %w: limit or offset cannot be zero", ErrInvalidArg)
	}

	visitorSelect := visitorRepo.visitorSelect
	if onlyActive {
		visitorSelect = visitorSelect.Where(visitorRepo.whereActive)
	}

	if residentID != "" {
		visitorSelect = visitorSelect.Where("visitor.resident_id = $1", residentID)
	}

	if search != "" {
		visitorSelect = visitorSelect.Where(squirrel.Or{
			squirrel.Expr("visitor.resident_id ILIKE $1", "%"+search+"%"),
			squirrel.Expr("visitor.first_name ILIKE $1"),
			squirrel.Expr("visitor.last_name ILIKE $1"),
		})
	}

	query, args, err := visitorSelect.
		Limit(uint64(getBoundedLimit(limit))).
		Offset(uint64(offset)).
		OrderBy("visitor.id ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("visitor_repo.Get: %w: %v", ErrBuildingQuery, err)
	}

	visitors := visitorSlice{}
	err = visitorRepo.database.driver.Select(&visitors, query, args...)
	if err != nil {
		return nil, fmt.Errorf("visitor_repo.Get: %w: %v", ErrDatabaseQuery, err)
	}

	return visitors.toModels(), nil
}

func (visitorRepo VisitorRepo) GetCount(onlyActive bool, residentID string) (int, error) {
	countSelect := squirrel.Select("count(*)").From("visitor")
	if onlyActive {
		countSelect = countSelect.Where(visitorRepo.whereActive)
	}

	if residentID != "" {
		countSelect = countSelect.Where("visitor.resident_id = $1", residentID)
	}

	query, args, err := countSelect.ToSql()
	if err != nil {
		return 0, fmt.Errorf("visitor_repo.GetCount: %w: %v", ErrBuildingQuery, err)
	}

	var totalAmount int
	err = visitorRepo.database.driver.Get(&totalAmount, query, args...)
	if err != nil {
		return 0, fmt.Errorf("visitor_repo.GetCount: %w: %v", ErrDatabaseQuery, err)
	}

	return totalAmount, nil
}

func (visitorRepo VisitorRepo) Create(desiredVisitor models.Visitor) (string, error) {
	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := sq.
		Insert("visitor").
		SetMap(squirrel.Eq{
			"id":           desiredVisitor.ID,
			"resident_id":  desiredVisitor.ResidentID,
			"first_name":   desiredVisitor.FirstName,
			"last_name":    desiredVisitor.LastName,
			"relationship": desiredVisitor.Relationship,
			"access_start": desiredVisitor.AccessStart,
			"access_end":   desiredVisitor.AccessEnd,
		}).
		Suffix("RETURNING visitor.id").
		ToSql()
	if err != nil {
		return "", fmt.Errorf("visitor_repo.Create: %w: %v", ErrBuildingQuery, err)
	}

	var visitorID string
	err = visitorRepo.database.driver.Get(&visitorID, query, args...)
	if err != nil {
		return "", fmt.Errorf("visitor_repo.Create: %w: %v", ErrDatabaseExec, err)
	}

	return visitorID, nil
}

func (visitorRepo VisitorRepo) Delete(visitorID string) error {
	if visitorID == "" {
		return fmt.Errorf("visitor_repo.Delete: %w: negative or zero ID argument", ErrInvalidArg)
	}
	const query = `DELETE FROM visitor WHERE id = $1`

	res, err := visitorRepo.database.driver.Exec(query, visitorID)
	if err != nil {
		return fmt.Errorf("visitor_repo.Delete: %w: %v", ErrDatabaseExec, err)
	}

	if rowsAffected, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("visitor_repo.Delete: %w: %v", ErrGetRowsAffected, err)
	} else if rowsAffected == 0 {
		return fmt.Errorf("visitor_repo.Delete: %w", ErrNoRows)
	}

	return nil
}

func (visitorRepo VisitorRepo) GetOne(visitorID string) (models.Visitor, error) {
	if visitorID == "" {
		return models.Visitor{}, fmt.Errorf("visitor_repo.GetOne: %w: Empty ID argument", ErrInvalidArg)
	}

	query, args, err := visitorRepo.visitorSelect.Where("visitor.id = $1", visitorID).ToSql()
	if err != nil {
		return models.Visitor{}, fmt.Errorf("visitor_repo.GetOne: %w: %v", ErrBuildingQuery, err)
	}

	visitor := visitor{}
	err = visitorRepo.database.driver.Get(&visitor, query, args...)
	if err == sql.ErrNoRows {
		return models.Visitor{}, fmt.Errorf("visitor_repo.GetOne: %w", ErrNoRows)
	} else if err != nil {
		return models.Visitor{}, fmt.Errorf("visitor_repo.GetOne: %w: %v", ErrDatabaseQuery, err)
	}

	return visitor.toModels(), nil
}
