package main

import (
	"fmt"
	"github.com/dannyvelas/parkspot-api/config"
	"github.com/dannyvelas/parkspot-api/response"
	"github.com/dannyvelas/parkspot-api/storage"
	_ "github.com/joho/godotenv/autoload"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func main() {
	config, err := config.New()
	if err != nil {
		panic(err)
	}

	database, err := storage.NewDatabase(config.Postgres)
	if err != nil {
		panic(err)
	}

	permitRepo := storage.NewPermitRepo(database)

	router := httprouter.New()
	router.HandlerFunc("/api/permits", response.PermitsRouter(permitRepo))

	log.Fatal(http.ListenAndServe(":5000", router))
}
