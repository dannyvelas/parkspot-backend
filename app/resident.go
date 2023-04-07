package app

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/dannyvelas/lasvistas_api/storage/selectopts"
	"github.com/dannyvelas/lasvistas_api/util"
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
	opts := []selectopts.SelectOpt{
		selectopts.WithLimitAndOffset(boundedLimit, offset),
		selectopts.WithSearch(search),
	}

	allResidents, err := s.residentRepo.SelectWhere(models.Resident{}, opts...)
	if err != nil {
		return models.ListWithMetadata[models.Resident]{}, fmt.Errorf("resident_service.getAll: Error querying residentRepo: %v", err)
	}
	allResidents = util.MapSlice(allResidents, s.removeHash)

	totalAmount, err := s.residentRepo.SelectCountWhere(models.Resident{}, opts...)
	if err != nil {
		return models.ListWithMetadata[models.Resident]{}, fmt.Errorf("resident_service.getAll: Error getting total amount: %v", err)
	}

	return models.NewListWithMetadata(allResidents, totalAmount), nil
}

func (s ResidentService) GetOne(id string) (models.Resident, error) {
	if id == "" {
		return models.Resident{}, errs.MissingIDField
	}
	residents, err := s.residentRepo.SelectWhere(models.Resident{ID: id})
	if err != nil {
		return models.Resident{}, err
	} else if len(residents) == 0 {
		return models.Resident{}, errs.NewNotFound("resident")
	}
	resident := residents[0]

	return s.removeHash(resident), nil
}

func (s ResidentService) Update(desiredResident models.Resident) (models.Resident, error) {
	// if a password is being changed, make sure it is hashed before setting it in db
	if desiredResident.Password != "" {
		hashBytes, err := bcrypt.GenerateFromPassword([]byte(desiredResident.Password), bcrypt.DefaultCost)
		if err != nil {
			return models.Resident{}, fmt.Errorf("residentService.Update: error generating hash for password: %v", err)
		}
		desiredResident.Password = string(hashBytes)
	}

	err := s.residentRepo.Update(desiredResident)
	if err != nil {
		return models.Resident{}, fmt.Errorf("residentService.Update: Error updating resident: %w", err)
	}

	resident, err := s.GetOne(desiredResident.ID)
	if err != nil {
		return models.Resident{}, err
	}

	return s.removeHash(resident), nil
}

func (s ResidentService) Delete(id string) error {
	if id == "" {
		return errs.MissingIDField
	}

	return s.residentRepo.Delete(id)
}

func (s ResidentService) Create(desiredRes models.Resident) error {
	if err := desiredRes.ValidateCreation(); err != nil {
		return err
	}

	if residents, err := s.residentRepo.SelectWhere(models.Resident{ID: desiredRes.ID}); err != nil {
		return fmt.Errorf("resident_service.createResident: error getting resident by id: %v", err)
	} else if len(residents) != 0 {
		return errs.AlreadyExists("resident with ID " + desiredRes.ID)
	}

	if residents, err := s.residentRepo.SelectWhere(models.Resident{Email: desiredRes.Email}); err != nil {
		return fmt.Errorf("resident_service.createResident error getting resident by email: %v", err)
	} else if len(residents) != 0 {
		return errs.AlreadyExists("a resident with this email")
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(desiredRes.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("resident_service.createResident: error generating hash:" + err.Error())
	}
	hashString := string(hashBytes)

	desiredRes.Password = hashString
	err = s.residentRepo.Create(desiredRes)
	if err != nil {
		return fmt.Errorf("resident_service.createResident: Error querying residentRepo: %v", err)
	}

	return nil
}

// helpers
func (s ResidentService) removeHash(resident models.Resident) models.Resident {
	newRes := resident
	newRes.Password = ""
	return newRes
}
