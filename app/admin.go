package app

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
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

	return s.removeHash(admin), nil
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

	return s.removeHash(admin), nil
}

// helpers
func (s AdminService) removeHash(admin models.Admin) models.Admin {
	newRes := admin
	newRes.Password = ""
	return newRes
}
