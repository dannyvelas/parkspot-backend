package api

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/storage"
	"net/http/httptest"
)

func newTestServer() (*httptest.Server, error) {
	config, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("Error loading config: %v", err.Error())
	}

	database, err := storage.NewDatabase(config.Postgres())
	if err != nil {
		return nil, fmt.Errorf("Failed to start database: %v", err)
	}

	// init repos
	adminRepo := storage.NewAdminRepo(database)
	permitRepo := storage.NewPermitRepo(database)
	carRepo := storage.NewCarRepo(database)
	residentRepo := storage.NewResidentRepo(database)
	visitorRepo := storage.NewVisitorRepo(database)

	// http setup
	httpConfig := config.Http()

	router := NewRouter(httpConfig, config.Token(), config.OAuth(), config.Constants().DateFormat(),
		adminRepo, permitRepo, carRepo, residentRepo, visitorRepo)

	testServer := httptest.NewServer(router)

	return testServer, nil
}
