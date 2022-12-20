package app

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/rs/zerolog/log"
)

type App struct {
	JWTService      JWTService
	AuthService     AuthService
	ResidentService ResidentService
	VisitorService  VisitorService
	CarService      CarService
	PermitService   PermitService
}

func NewApp(c config.Config) (App, error) {
	// connect to database
	// no defer close() because connection closes automatically on program exit
	database, err := storage.NewDatabase(c.Postgres())
	if err != nil {
		return App{}, fmt.Errorf("Failed to start database: %v", err)
	}
	log.Info().Msg("Connected to Database.")

	// repos
	adminRepo := storage.NewAdminRepo(database)
	residentRepo := storage.NewResidentRepo(database)
	visitorRepo := storage.NewVisitorRepo(database)
	carRepo := storage.NewCarRepo(database)
	permitRepo := storage.NewPermitRepo(database)

	// services
	jwtService := NewJWTService(c.Token())
	authService := NewAuthService(jwtService, adminRepo, residentRepo, c.Http(), c.OAuth())
	residentService := NewResidentService(residentRepo)
	visitorService := NewVisitorService(visitorRepo)
	carService := NewCarService(carRepo)
	permitService := NewPermitService(permitRepo, residentRepo, carRepo)

	return App{
		JWTService:      jwtService,
		AuthService:     authService,
		ResidentService: residentService,
		VisitorService:  visitorService,
		CarService:      carService,
		PermitService:   permitService,
	}, nil
}
