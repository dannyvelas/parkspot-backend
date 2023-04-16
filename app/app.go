package app

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/storage/psql"
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
	database, err := psql.NewDatabase(c.Postgres())
	if err != nil {
		return App{}, fmt.Errorf("Failed to start database: %v", err)
	}
	log.Info().Msg("Connected to Database.")

	// repos
	adminRepo := psql.NewAdminRepo(database)
	residentRepo := psql.NewResidentRepo(database)
	visitorRepo := psql.NewVisitorRepo(database)
	carRepo := psql.NewCarRepo(database)
	permitRepo := psql.NewPermitRepo(database)

	// services
	jwtService := NewJWTService(c.Token())
	adminService := NewAdminService(adminRepo)
	residentService := NewResidentService(residentRepo)
	authService := NewAuthService(jwtService, adminService, residentService, c.Http(), c.OAuth())
	visitorService := NewVisitorService(visitorRepo)
	carService := NewCarService(carRepo)
	permitService := NewPermitService(permitRepo, residentRepo, carService)

	return App{
		JWTService:      jwtService,
		AuthService:     authService,
		ResidentService: residentService,
		VisitorService:  visitorService,
		CarService:      carService,
		PermitService:   permitService,
	}, nil
}
