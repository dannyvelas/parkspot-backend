package main

import (
	"github.com/dannyvelas/parkspot-api/auth"
	"github.com/dannyvelas/parkspot-api/config"
	"github.com/dannyvelas/parkspot-api/routing"
	"github.com/dannyvelas/parkspot-api/storage"
	"github.com/go-chi/chi/v5"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Info().Msg("Initializing app...")

	config, err := config.New()
	if err != nil {
		log.Fatal().Msgf("Failed loading config: %s", err.Error())
		return
	}

	database, err := storage.NewDatabase(config.Postgres())
	if err != nil {
		log.Fatal().Msgf("Failed to start database: %s", err.Error())
		return
	}
	log.Info().Msg("Connected to Database.")

	adminRepo := storage.NewAdminRepo(database)
	permitRepo := storage.NewPermitRepo(database)

	authenticator := auth.NewAuthenticator(config.Token())

	router := chi.NewRouter()
	router.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Post("/login", routing.Login(authenticator, adminRepo))
		apiRouter.Route("/admin", func(adminRouter chi.Router) {
			//adminRouter.Use(authenticator.AdminOnly)
			adminRouter.Route("/permits", routing.PermitsRouter(permitRepo))
		})
	})

	err = http.ListenAndServe(":5000", router)
	if err != nil {
		log.Fatal().Msgf("Failed to start server: %s", err.Error())
	}
}
