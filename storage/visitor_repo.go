package storage

import (
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/lasvistas_api/models"
)

type VisitorRepo struct {
	database      Database
	visitorSelect squirrel.SelectBuilder
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

	return VisitorRepo{database: database, visitorSelect: visitorSelect}
}

func (visitorRepo VisitorRepo) GetAll(limit, offset int) ([]models.Visitor, error) {
	if limit < 0 || offset < 0 {
		return nil, fmt.Errorf("visitor_repo.GetAll: %w: limit or offset cannot be zero", ErrInvalidArg)
	}

	query, _, err := visitorRepo.visitorSelect.
		Limit(uint64(getBoundedLimit(limit))).
		Offset(uint64(offset)).
		OrderBy("visitor.id ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("visitor_repo.GetAll: %w: %v", ErrBuildingQuery, err)
	}

	visitors := visitorSlice{}
	err = visitorRepo.database.driver.Select(&visitors, query)
	if err != nil {
		return nil, fmt.Errorf("visitor_repo.GetAll: %w: %v", ErrDatabaseQuery, err)
	}

	return visitors.toModels(), nil
}

func (visitorRepo VisitorRepo) GetAllTotalAmount() (int, error) {
	const query = "SELECT count(*) FROM visitor"

	var totalAmount int
	err := visitorRepo.database.driver.Get(&totalAmount, query)
	if err != nil {
		return 0, fmt.Errorf("visitor_repo.GetAllTotalAmount: %w: %v", ErrDatabaseQuery, err)
	}

	return totalAmount, nil
}
