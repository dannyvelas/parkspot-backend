package app

import (
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"golang.org/x/crypto/bcrypt"
)

type ResidentService struct {
	residentRepo storage.ResidentRepo
}

func NewResidentService(residentRepo storage.ResidentRepo) ResidentService {
	return ResidentService{
		residentRepo: residentRepo,
	}
}

func (s ResidentService) GetAll(limit, page int, search string) (models.ListWithMetadata[models.Resident], error) {
	boundedLimit, offset := getBoundedLimitAndOffset(limit, page)

	allResidents, err := s.residentRepo.GetAll(boundedLimit, offset, search)
	if err != nil {
		return models.ListWithMetadata[models.Resident]{}, fmt.Errorf("resident_service.getAll: Error querying residentRepo: %v", err)
	}

	totalAmount, err := s.residentRepo.GetAllTotalAmount()
	if err != nil {
		return models.ListWithMetadata[models.Resident]{}, fmt.Errorf("resident_service.getAll: Error getting total amount: %v", err)
	}

	return models.NewListWithMetadata(allResidents, totalAmount), nil
}

func (s ResidentService) GetOne(id string) (models.Resident, error) {
	resident, err := s.residentRepo.GetOne(id)
	if err != nil && !errors.Is(err, storage.ErrNoRows) {
		return models.Resident{}, fmt.Errorf("resident_service.getOne: Error getting resident: %v", err)
	} else if errors.Is(err, storage.ErrNoRows) {
		return models.Resident{}, ErrNotFound
	}

	return resident, nil
}

func (s ResidentService) Update(id string, desiredResident models.EditResident) (models.Resident, error) {
	err := s.residentRepo.Update(id, desiredResident)
	if err != nil {
		return models.Resident{}, fmt.Errorf("resident_service.editResident: Error updating resident: %v", err)
	}

	resident, err := s.residentRepo.GetOne(id)
	if err != nil {
		return models.Resident{}, fmt.Errorf("resident_service.editResident: Error getting resident: %v", err)
	}

	return resident, nil
}

func (s ResidentService) Delete(id string) error {
	resident, err := s.residentRepo.GetOne(id)
	if errors.Is(err, storage.ErrNoRows) {
		return ErrNotFound
	} else if err != nil {
		return fmt.Errorf("resident_service.deleteResident: Error getting resident: %v", err)
	}

	err = s.residentRepo.Delete(resident.Id)
	if errors.Is(err, storage.ErrNoRows) {
		return ErrNotFound
	} else if err != nil {
		return fmt.Errorf("resident_service.deleteResident: %v", err)
	}

	return nil
}

func (s ResidentService) Create(desiredRes models.CreateResident) error {
	if _, err := s.residentRepo.GetOne(desiredRes.ResidentId); err == nil {
		return newErrAlreadyExists("resident with ID " + desiredRes.ResidentId)
	} else if !errors.Is(err, storage.ErrNoRows) {
		return fmt.Errorf("resident_service.createResident: error getting resident by id: %v", err)
	}

	if _, err := s.residentRepo.GetOneByEmail(desiredRes.Email); err == nil {
		return newErrAlreadyExists("resident with this email")
	} else if !errors.Is(err, storage.ErrNoRows) {
		return fmt.Errorf("resident_service.createResident error getting resident by email: %v", err)
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(desiredRes.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("resident_service.createResident: error generating hash:" + err.Error())
	}
	hashString := string(hashBytes)

	resident := models.NewResident(desiredRes.ResidentId,
		desiredRes.FirstName,
		desiredRes.LastName,
		desiredRes.Phone,
		desiredRes.Email,
		hashString,
		desiredRes.UnlimDays,
		0, 0)

	err = s.residentRepo.Create(resident)
	if err != nil {
		return fmt.Errorf("resident_service.createResident: Error querying residentRepo: %v", err)
	}

	return nil
}
