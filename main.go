package main

import (
	"context"
	"fmt"
	"github.com/dannyvelas/parkspot-api/auth"
	"github.com/dannyvelas/parkspot-api/config"
	"github.com/dannyvelas/parkspot-api/routing"
	"github.com/dannyvelas/parkspot-api/storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Info().Msg("Initializing app...")

	// load config
	config, err := config.New()
	if err != nil {
		log.Fatal().Msgf("Failed loading config: %s", err)
		return
	}

	// connect to database
	database, err := storage.NewDatabase(config.Postgres())
	if err != nil {
		log.Fatal().Msgf("Failed to start database: %s", err)
		return
	}
	log.Info().Msg("Connected to Database.")

	// initialize repos
	adminRepo := storage.NewAdminRepo(database)
	permitRepo := storage.NewPermitRepo(database)

	// initialize authenticator
	authenticator := auth.NewAuthenticator(config.Token())

	// set routes
	router := chi.NewRouter()
	router.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Post("/login", routing.Login(authenticator, adminRepo))
		apiRouter.Route("/admin", func(adminRouter chi.Router) {
			//adminRouter.Use(authenticator.AdminOnly)
			adminRouter.Route("/permits", routing.PermitsRouter(permitRepo))
		})
	})

	// configure http server
	httpConfig := config.Http()
	httpServer := http.Server{
		Addr:         fmt.Sprintf(":%d", httpConfig.Port()),
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
		errChannel <- StartServer(httpServer)
	}()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChannel <- fmt.Errorf("%s", <-c)
	}()

	fatalErr := <-errChannel
	log.Info().Msgf("Closing server: %s", fatalErr)

	shutdownGracefully(30*time.Second, httpServer)
}

func StartServer(httpServer http.Server) error {
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatal().Msgf("Failed to start server: %s", err)
		return err
	}
	return nil
}

func shutdownGracefully(timeout time.Duration, httpServer http.Server) {
	log.Info().Msg("Gracefully shutting down...")

	gracefullCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := httpServer.Shutdown(gracefullCtx); err != nil {
		log.Error().Msgf("Error shutting down the server: %s", err)
	} else {
		log.Info().Msg("HttpServer gracefully shut down")
	}
}
