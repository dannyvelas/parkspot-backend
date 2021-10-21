package main

import (
	"github.com/dannyvelas/parkspot-api/config"
	"github.com/dannyvelas/parkspot-api/routing"
	"github.com/dannyvelas/parkspot-api/storage"
	"github.com/go-chi/chi/v5"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"net/http"
)

func main() {
	config, err := config.New()
	if err != nil {
		panic(err)
	}

	database, err := storage.NewDatabase(config.Postgres())
	if err != nil {
		panic(err)
	}

	adminRepo := storage.NewAdminRepo(database)
	permitRepo := storage.NewPermitRepo(database)

	router := chi.NewRouter()
	router.Post("/api/login", routing.HandleLogin(*adminRepo))
	router.Route("/api/permits", routing.PermitsRouter(*permitRepo))

	log.Fatal(http.ListenAndServe(":5000", router))
}
