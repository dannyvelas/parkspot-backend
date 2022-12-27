package app

import (
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
)

type VisitorService struct {
	visitorRepo storage.VisitorRepo
}

func NewVisitorService(visitorRepo storage.VisitorRepo) VisitorService {
	return VisitorService{
		visitorRepo: visitorRepo,
	}
}

func (s VisitorService) GetActive(limit, page int, search string, residentID string) (models.ListWithMetadata[models.Visitor], error) {
	boundedLimit, offset := getBoundedLimitAndOffset(limit, page)

	allVisitors, err := s.visitorRepo.Get(true, residentID, search, boundedLimit, offset)
	if err != nil {
		return models.ListWithMetadata[models.Visitor]{}, fmt.Errorf("error getting all visitors from visitor repo: %v", err)
	}

	totalAmount, err := s.visitorRepo.GetCount(true, residentID)
	if err != nil {
		return models.ListWithMetadata[models.Visitor]{}, fmt.Errorf("error getting count of all visitors from visitor repo: %v", err)
	}

	return models.NewListWithMetadata(allVisitors, totalAmount), nil
}

func (s VisitorService) Create(residentID string, desiredVisitor models.Visitor) (models.Visitor, error) {
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
	err := s.visitorRepo.Delete(id)
	if errors.Is(err, storage.ErrNoRows) {
		return errs.NotFound
	} else if err != nil {
		return fmt.Errorf("error deleting in visitorRepo: %v", err)
	}

	return nil
}
