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

	// load config
	config := config.NewConfig()

	// connect to database
	// no defer close() because connection closes automatically on program exit
	database, err := storage.NewDatabase(config.Postgres())
	if err != nil {
		log.Fatal().Msgf("Failed to start database: %v", err)
	}
	log.Info().Msg("Connected to Database.")

	// init repos
	adminRepo := storage.NewAdminRepo(database)
	permitRepo := storage.NewPermitRepo(database)
	carRepo := storage.NewCarRepo(database)
	residentRepo := storage.NewResidentRepo(database)

	// http setup
	httpConfig := config.Http()

	router := api.NewRouter(httpConfig, config.Token(), config.Constants().DateFormat(),
		adminRepo, permitRepo, carRepo, residentRepo)

	httpServer := http.Server{
		Addr:         httpConfig.Host() + ":" + httpConfig.Port(),
		Handler:      router,
		ReadTimeout:  httpConfig.ReadTimeout(),
		WriteTimeout: httpConfig.WriteTimeout(),
		IdleTimeout:  httpConfig.IdleTimeout(),
	}

	// initialize error channel
	errChannel := make(chan error)
	defer close(errChannel)

	// receive errors from startup or signal interrupt
	go func() {
		errChannel <- startServer(httpServer)
	}()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChannel <- fmt.Errorf("%s", <-c)
	}()

	fatalErr := <-errChannel
	log.Info().Msgf("Closing server: %v", fatalErr)

	shutdownGracefully(30*time.Second, httpServer)
}

func startServer(httpServer http.Server) error {
	log.Info().Msgf("Server started on: %s", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("Failed to start server: %v", err)
	}
	return nil
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
