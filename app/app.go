package app

import (
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/storage"
)

type App struct {
	JWTService      JWTService
	AuthService     AuthService
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
		ResidentService: residentService,
		VisitorService:  visitorService,
		CarService:      carService,
		PermitService:   permitService,
	}
}
