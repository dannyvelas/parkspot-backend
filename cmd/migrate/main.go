// inspired from: https://github.com/doppelganger113/simple_gopher/blob/main/cmd/migrate/main.go
package main

import (
	"flag"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/storage/psql"
	"github.com/dannyvelas/lasvistas_api/util"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

type settings struct {
	// if `command` is "up" or "down", this field represents the amount of steps to take
	// if `command` is "force", this field represents the version to migrate to
	// if `command` is "force", this field must be non-nil.
	// if `command` is "up" or "down", this field can be nil. In such case, a full up or down
	// migration will happen, respectively
	steps *int

	command      string
	databaseURL  string
	migrationDir string
}

func main() {
	settings, err := getSettings()
	if err != nil {
		log.Fatal().Msgf("%v", err)
	}

	migrator, err := initMigrator(settings)
	if err != nil {
		log.Fatal().Msgf("error initializing migrator: %v", err)
	}
	defer func() {
		if _, err := migrator.Close(); err != nil {
			log.Info().Msgf("error closing: %v", err)
		}
	}()

	// handle Ctrl+c
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT)
	go func() {
		for range signals {
			log.Info().Msg("Stopping after this running migration ...")
			migrator.GracefulStop <- true
			return
		}
	}()

	if err := execCommand(migrator, settings); err != nil {
		log.Fatal().Msgf("error executing command: %v", err)
	}

	log.Info().Msgf("successfully executed command: %s", settings.command)
}

func initMigrator(settings settings) (*migrate.Migrate, error) {
	database, err := psql.NewDatabase(config.PostgresConfig{URL: settings.databaseURL})
	if err != nil {
		return nil, fmt.Errorf("Failed to start database: %v", err)
	}
	log.Info().Msg("Connected to Database.")

	driver, err := postgres.WithInstance(database.Driver.DB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("Call to postgres.WithInstance failed to cast *sql.DB to migrate.Driver: %v", err)
	}

	sourceURL := fmt.Sprintf("file://%s", settings.migrationDir)
	migrator, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize migrate with migrate.Driver instance: %v", err)
	}

	return migrator, nil
}

func getSettings() (settings, error) {
	// required arguments
	databaseURL := flag.String("database", "", "the URL of the database that should experience the migration")
	migrationsDir := flag.String("path", "", "the directory where the migrations file exist")
	flag.Parse()

	if *databaseURL == "" {
		return settings{}, fmt.Errorf("error: databaseURL argument not provided")
	} else if *migrationsDir == "" {
		return settings{}, fmt.Errorf("error: migrationsDir argument not provided")
	}

	command := flag.Arg(0)
	if command == "" {
		return settings{}, fmt.Errorf("error: no subcommand")
	}

	retVal := settings{
		databaseURL:  *databaseURL,
		migrationDir: *migrationsDir,
		command:      command,
	}

	// `args` should hold every argument after the `command` argument
	// for example, for the following flags:
	// `go run cmd/migrate/main.go -path migrations -database $(DATABASE_URL) down 1`
	// `args` should be equal to ["1"]
	args := flag.Args()[1:]
	flagSet := flag.NewFlagSet(command, flag.ExitOnError)
	if flagSet.NArg() == 0 && command == "force" {
		return settings{}, fmt.Errorf("error: need <version> argument to \"force\" subcommand")
	}

	if err := flagSet.Parse(args); err != nil {
		return settings{}, fmt.Errorf("error: need <version> argument to \"force\" subcommand")
	}

	if flagSet.NArg() > 0 {
		n, err := strconv.ParseUint(flagSet.Arg(0), 10, 64)
		if err != nil {
			return settings{}, fmt.Errorf("error: can't read limit argument N")
		}
		retVal.steps = util.ToPtr(int(n))
	}

	return retVal, nil
}

func execCommand(migrator *migrate.Migrate, settings settings) (err error) {
	switch settings.command {
	case "force":
		err = migrator.Force(*settings.steps)
	case "up":
		if settings.steps != nil {
			err = migrator.Steps(*settings.steps)
		} else {
			err = migrator.Up()
		}
	case "down":
		if settings.steps != nil {
			err = migrator.Steps(-(*settings.steps))
		} else {
			err = migrator.Down()
		}
	case "version":
		var v uint
		var dirty bool
		v, dirty, err = migrator.Version()
		if dirty {
			log.Printf("%v (dirty)\n", v)
		} else {
			log.Printf("%v\n", v)
		}
	default:
		return fmt.Errorf("error: unrecognized subcommand")
	}
	if err != migrate.ErrNoChange {
		return err
	}
	return nil
}
