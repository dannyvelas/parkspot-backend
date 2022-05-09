package main

import (
	"context"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/api"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// logging setup
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Info().Msg("Initializing app...")

	// set global values
	const dateFormat = "2006-01-02"

	// load config
	config := config.NewConfig()

	/************************************** DATABASE *********************************/
	// connect to database
	// no defer close() because connection closes automatically on program exit
	database, err := storage.NewDatabase(config.Postgres())
	if err != nil {
		log.Fatal().Msgf("Failed to start database: %v", err)
	}
	log.Info().Msg("Connected to Database.")

	adminRepo := storage.NewAdminRepo(database)
	permitRepo := storage.NewPermitRepo(database)
	carRepo := storage.NewCarRepo(database)
	residentRepo := storage.NewResidentRepo(database)
	/************************************** DATABASE *********************************/

	/************************************** HTTP *********************************/
	httpConfig := config.Http()

	router := api.NewRouter(httpConfig, config.Token(), dateFormat,
		adminRepo, permitRepo, carRepo, residentRepo)

	httpServer := http.Server{
		Addr:         fmt.Sprintf("%s:%d", httpConfig.Host(), httpConfig.Port()),
		Handler:      router,
		ReadTimeout:  httpConfig.ReadTimeout(),
		WriteTimeout: httpConfig.WriteTimeout(),
		IdleTimeout:  httpConfig.IdleTimeout(),
	}
	/************************************** HTTP *********************************/

	// initialize error channel
	errChannel := make(chan error)
	defer close(errChannel)

	// receive errors from startup or signal interrupt
	go StartServer(httpServer)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChannel <- fmt.Errorf("%s", <-c)
	}()

	fatalErr := <-errChannel
	log.Info().Msgf("Closing server: %v", fatalErr)

	shutdownGracefully(30*time.Second, httpServer)
}

func StartServer(httpServer http.Server) {
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatal().Msgf("Failed to start server: %v", err)
	}
}

func shutdownGracefully(timeout time.Duration, httpServer http.Server) {
	log.Info().Msg("Gracefully shutting down...")

	gracefullCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := httpServer.Shutdown(gracefullCtx); err != nil {
		log.Error().Msgf("Error shutting down the server: %v", err)
	} else {
		log.Info().Msg("HttpServer gracefully shut down")
	}
}
