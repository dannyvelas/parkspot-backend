package app

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/models/validator"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/dannyvelas/lasvistas_api/storage/selectopts"
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

	allResidents, err := s.residentRepo.SelectWhere(models.Resident{},
		selectopts.WithLimitAndOffset(boundedLimit, offset),
		selectopts.WithSearch(search),
	)
	if err != nil {
		return models.ListWithMetadata[models.Resident]{}, fmt.Errorf("resident_service.getAll: Error querying residentRepo: %v", err)
	}

	totalAmount, err := s.residentRepo.SelectCountWhere(models.Resident{}, selectopts.WithSearch(search))
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

	return resident, nil
}

func (s ResidentService) Update(desiredResident models.Resident) (models.Resident, error) {
	if desiredResident.ID == "" {
		return models.Resident{}, errs.MissingIDField
	}
	// this check goes here; not in `validator.EditResident` bc this err is mut. exclusive w those errs
	if desiredResident.FirstName == "" && desiredResident.LastName == "" &&
		desiredResident.Phone == "" && desiredResident.Email == "" &&
		desiredResident.UnlimDays == nil && desiredResident.AmtParkingDaysUsed == nil {
		return models.Resident{}, errs.AllEditFieldsEmpty("firstName, lastName, phone, email, unlimDays, amtParkingDaysUsed")
	}

	if err := validator.EditResident.Run(desiredResident); err != nil {
		return models.Resident{}, err
	}

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

	return resident, nil
}

func (s ResidentService) Delete(id string) error {
	if id == "" {
		return errs.MissingIDField
	}

	return s.residentRepo.Delete(id)
}

func (s ResidentService) Create(desiredRes models.Resident) (models.Resident, error) {
	if err := validator.CreateResident.Run(desiredRes); err != nil {
		return models.Resident{}, err
	}

	if residents, err := s.residentRepo.SelectWhere(models.Resident{ID: desiredRes.ID}); err != nil {
		return models.Resident{}, fmt.Errorf("resident_service.createResident: error getting resident by id: %v", err)
	} else if len(residents) != 0 {
		return models.Resident{}, errs.AlreadyExists("a resident with ID: " + desiredRes.ID)
	}

	if residents, err := s.residentRepo.SelectWhere(models.Resident{Email: desiredRes.Email}); err != nil {
		return models.Resident{}, fmt.Errorf("resident_service.createResident error getting resident by email: %v", err)
	} else if len(residents) != 0 {
		return models.Resident{}, errs.AlreadyExists("a resident with this email: " + desiredRes.Email)
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(desiredRes.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.Resident{}, fmt.Errorf("resident_service.createResident: error generating hash:" + err.Error())
	}
	hashString := string(hashBytes)

	desiredRes.Password = hashString
	err = s.residentRepo.Create(desiredRes)
	if err != nil {
		return models.Resident{}, fmt.Errorf("resident_service.createResident: Error querying residentRepo: %v", err)
	}

	return desiredRes, nil
}
