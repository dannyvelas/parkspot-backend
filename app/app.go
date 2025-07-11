package app

import (
	"github.com/dannyvelas/parkspot-backend/config"
	"github.com/dannyvelas/parkspot-backend/storage"
)

type App struct {
	JWTService      JWTService
	AuthService     AuthService
	AdminService    AdminService
	ResidentService ResidentService
	VisitorService  VisitorService
	CarService      CarService
	PermitService   PermitService
}

func NewApp(c config.Config, database storage.Database) App {
	// services
	jwtService := NewJWTService(c.Token)
	adminService := NewAdminService(database.AdminRepo())
	residentService := NewResidentService(database.ResidentRepo())
	authService := NewAuthService(jwtService, adminService, residentService, c.Http, c.OAuth)
	visitorService := NewVisitorService(database.VisitorRepo())
	carService := NewCarService(database.CarRepo())
	permitService := NewPermitService(database.PermitRepo(), database.ResidentRepo(), carService)

	return App{
		JWTService:      jwtService,
		AuthService:     authService,
		AdminService:    adminService,
		ResidentService: residentService,
		VisitorService:  visitorService,
		CarService:      carService,
		PermitService:   permitService,
	}
}
