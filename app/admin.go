package app

import (
	"errors"
	"fmt"

	"github.com/dannyvelas/parkspot-backend/errs"
	"github.com/dannyvelas/parkspot-backend/models"
	"github.com/dannyvelas/parkspot-backend/models/validator"
	"github.com/dannyvelas/parkspot-backend/storage"
	"golang.org/x/crypto/bcrypt"
)

type AdminService struct {
	adminRepo storage.AdminRepo
}

func NewAdminService(adminRepo storage.AdminRepo) AdminService {
	return AdminService{
		adminRepo: adminRepo,
	}
}

func (s AdminService) GetOne(id string) (models.Admin, error) {
	if id == "" {
		return models.Admin{}, errs.MissingIDField
	}

	admin, err := s.adminRepo.GetOne(id)
	if err != nil {
		return models.Admin{}, err
	}

	return admin, nil
}

func (s AdminService) Update(desiredAdmin models.Admin) (models.Admin, error) {
	// if a password is being changed, make sure it is hashed before setting it in db
	if desiredAdmin.Password != "" {
		hashBytes, err := bcrypt.GenerateFromPassword([]byte(desiredAdmin.Password), bcrypt.DefaultCost)
		if err != nil {
			return models.Admin{}, fmt.Errorf("adminService.Update: error generating hash for password: %v", err)
		}
		desiredAdmin.Password = string(hashBytes)
	}

	err := s.adminRepo.Update(desiredAdmin)
	if err != nil {
		return models.Admin{}, fmt.Errorf("adminService.Update: Error updating admin: %w", err)
	}

	admin, err := s.GetOne(desiredAdmin.ID)
	if err != nil {
		return models.Admin{}, err
	}

	return admin, nil
}

func (s AdminService) Create(desiredAdmin models.Admin) (models.Admin, error) {
	if err := validator.CreateAdmin.Run(desiredAdmin); err != nil {
		return models.Admin{}, err
	}

	// make sure admin doesn't already exist
	if _, err := s.adminRepo.GetOne(desiredAdmin.ID); err != nil && !errors.Is(err, errs.NotFound) {
		return models.Admin{}, fmt.Errorf("admin_service.createAdmin: error getting admin by id: %v", err)
	} else if err == nil {
		return models.Admin{}, errs.NewAlreadyExists("an admin with ID: " + desiredAdmin.ID)
	}

	// TODO: check for duplicate email here

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(desiredAdmin.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.Admin{}, fmt.Errorf("admin_service.create: error generating hash: %v", err)
	}
	hashString := string(hashBytes)

	desiredAdmin.Password = hashString
	err = s.adminRepo.Create(desiredAdmin)
	if err != nil {
		return models.Admin{}, fmt.Errorf("admin_service.createAdmin: Error querying adminRepo: %v", err)
	}

	createdAdmin, err := s.GetOne(desiredAdmin.ID)
	if err != nil {
		return models.Admin{}, fmt.Errorf("error getting admin which was just created: %w", err)
	}

	return createdAdmin, nil
}

// not used yet but could come in handy
func (s AdminService) Delete(id string) error {
	if id == "" {
		return errs.MissingIDField
	}

	return s.adminRepo.Delete(id)
}
