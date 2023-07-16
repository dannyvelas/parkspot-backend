package main

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/api"
	"github.com/dannyvelas/lasvistas_api/app"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/storage/psql"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	c, err := config.NewConfig()
	if err != nil {
		log.Fatal().Msgf("Error loading config: %v", err)
	}

	// connect to database
	// no defer close() because connection closes automatically on program exit
	database, err := psql.NewDatabase(c.Postgres)
	if err != nil {
		log.Fatal().Msgf("Failed to start database: %v", err)
	}
	log.Info().Msg("Connected to Database.")

	// create app
	app := app.NewApp(c, database)

	// initialize error channel
	errChannel := make(chan error)
	defer close(errChannel)

	// initialize and start server
	server := api.NewServer(c, app)
	go server.Start(errChannel)

	// listen to signal interrupt
	go listenToInterrupt(errChannel)

	fatalErr := <-errChannel
	log.Info().Msgf("Closing server: %v", fatalErr)

	server.ShutdownGracefully(30 * time.Second)
}

func listenToInterrupt(errChannel chan<- error) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	errChannel <- fmt.Errorf("%s", <-c)
}
