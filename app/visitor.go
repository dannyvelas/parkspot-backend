package app

import (
	"fmt"
	"github.com/dannyvelas/parkspot-backend/errs"
	"github.com/dannyvelas/parkspot-backend/models"
	"github.com/dannyvelas/parkspot-backend/storage"
	"github.com/dannyvelas/parkspot-backend/storage/selectopts"
)

type VisitorService struct {
	visitorRepo storage.VisitorRepo
}

func NewVisitorService(visitorRepo storage.VisitorRepo) VisitorService {
	return VisitorService{
		visitorRepo: visitorRepo,
	}
}

func (s VisitorService) Get(status models.Status, limit, page int, search, residentID string) (models.ListWithMetadata[models.Visitor], error) {
	boundedLimit, offset := getBoundedLimitAndOffset(limit, page)

	allVisitors, err := s.visitorRepo.SelectWhere(models.Visitor{ResidentID: residentID},
		selectopts.WithStatus(status),
		selectopts.WithSearch(search),
		selectopts.WithLimitAndOffset(boundedLimit, offset),
	)
	if err != nil {
		return models.ListWithMetadata[models.Visitor]{}, fmt.Errorf("error getting all visitors from visitor repo: %v", err)
	}

	totalAmount, err := s.visitorRepo.SelectCountWhere(models.Visitor{ResidentID: residentID},
		selectopts.WithStatus(status),
		selectopts.WithSearch(search),
	)
	if err != nil {
		return models.ListWithMetadata[models.Visitor]{}, fmt.Errorf("error getting count of all visitors from visitor repo: %v", err)
	}

	return models.NewListWithMetadata(allVisitors, totalAmount), nil
}

func (s VisitorService) GetOne(id string) (models.Visitor, error) {
	if id == "" {
		return models.Visitor{}, errs.MissingIDField
	}
	return s.visitorRepo.GetOne(id)
}

func (s VisitorService) Create(desiredVisitor models.Visitor) (models.Visitor, error) {
	if err := desiredVisitor.ValidateCreation(); err != nil {
		return models.Visitor{}, err
	}

	visitorID, err := s.visitorRepo.Create(desiredVisitor)
	if err != nil {
		return models.Visitor{}, fmt.Errorf("error creating visitor in visitor repo: %v", err)
	}

	visitor, err := s.visitorRepo.GetOne(visitorID)
	if err != nil {
		return models.Visitor{}, fmt.Errorf("error getting visitor after creating in visitor repo: %v", err)
	}

	return visitor, nil
}

func (s VisitorService) Delete(id string) error {
	if id == "" {
		return errs.MissingIDField
	}
	return s.visitorRepo.Delete(id)
}
