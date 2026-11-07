package main

import (
	"github.com/dannyvelas/parkspot-backend/config"
	"github.com/dannyvelas/parkspot-backend/migrations"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	c, err := config.NewConfig()
	if err != nil {
		log.Fatal().Msgf("Error loading config: %v", err)
	}

	if err := migrations.Up(c.Postgres.URL); err != nil {
		log.Fatal().Msgf("Failed to apply migrations: %v", err)
	}

	log.Info().Msg("Migrations applied.")
}
