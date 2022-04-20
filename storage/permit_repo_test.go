package storage

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/golang-migrate/migrate/v4"
	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
	"time"
)

type permitRepoSuite struct {
	suite.Suite
	location   *time.Location
	permitRepo PermitRepo
	migrator   *migrate.Migrate
}

func TestPermitRepo(t *testing.T) {
	suite.Run(t, new(permitRepoSuite))
}

func (suite *permitRepoSuite) SetupSuite() {
	config := config.NewConfig()

	database, err := NewDatabase(config.Postgres())
	if err != nil {
		log.Fatal().Msgf("Failed to start database: %v", err)
	}
	suite.permitRepo = NewPermitRepo(database, "2006-01-02")

	migrator, err := GetV1Migrator(database)
	if err != nil {
		log.Fatal().Msgf("Failed to get migrator: %v", err)
	}
	suite.migrator = migrator

	if err := suite.migrator.Up(); err != nil {
		log.Fatal().Msgf("Error when migrating all the way up: %v", err)
	}
}

func (suite permitRepoSuite) TearDownSuite() {
	err := suite.migrator.Down()
	if err != nil {
		suite.NoError(err, "No error migrating all the way down")
	}
}

func (suite permitRepoSuite) TestGetAllPermits_EmptySlice_Positive() {
	err := suite.migrator.Migrate(1)
	suite.NoError(err, "No error when migrating down to v1")
	defer func() {
		err := suite.migrator.Up()
		suite.NoError(err, "No error when migrating all the way up again")
	}()

	permits, err := suite.permitRepo.GetAll(defaultLimit, defaultOffset)
	suite.NoError(err, "no error getting all permit when the table is empty")
	suite.Equal(0, len(permits), "length of permit should be 0")
	suite.True(cmp.Equal(permits, []models.Permit{}), "permit should be an empty slice")
}

func (suite permitRepoSuite) TestGetActivePermits_EmptySlice_Positive() {
	err := suite.migrator.Migrate(1)
	suite.NoError(err, "No error when migrating down to v1")
	defer func() {
		err := suite.migrator.Up()
		suite.NoError(err, "No error when migrating all the way up again")
	}()

	permits, err := suite.permitRepo.GetActive(defaultLimit, defaultOffset)
	suite.NoError(err, "no error getting active permits when the table is empty")
	suite.Equal(0, len(permits), "length of permits should be 0")
	suite.True(cmp.Equal(permits, []models.Permit{}), "permit should be an empty slice")
}

func (suite permitRepoSuite) TestGetAllPermits_NonEmpty_Positive() {
	// check that length is not 0
	permits, err := suite.permitRepo.GetAll(defaultLimit, defaultOffset)
	suite.NoError(err, "no error getting all permits when the table is not empty")
	suite.NotEqual(len(permits), 0, "length of permits should not be 0")
}

func (suite permitRepoSuite) TestGetActivePermits_NonEmpty_Positive() {
	// check that length is not 0
	permits, err := suite.permitRepo.GetActive(defaultLimit, defaultOffset)
	suite.NoError(err, "no error getting all permits when the table is not empty")
	suite.NotEqual(len(permits), 0, "length of permits should not be 0")
}

func (suite permitRepoSuite) TestWriteAllPermits_Positive() {
	permits, err := suite.permitRepo.GetAll(defaultLimit, defaultOffset)
	suite.NoError(err, "No error when getting all permits")

	f, err := os.Create("testout/all_permits.txt")
	suite.NoError(err, "No error creating all_permits file")
	defer f.Close()

	for _, permit := range permits {
		_, err := f.WriteString(permitToString(permit, suite.permitRepo.dateFormat))
		suite.NoError(err, "No error when writing line")
	}
}

func (suite permitRepoSuite) TestWriteActivePermits_Positive() {
	permits, err := suite.permitRepo.GetActive(defaultLimit, defaultOffset)
	suite.NoError(err, "No error when getting active permits")

	f, err := os.Create("testout/active_permits.txt")
	suite.NoError(err, "No error creating active_permits file")
	defer f.Close()

	for _, permit := range permits {
		_, err := f.WriteString(permitToString(permit, suite.permitRepo.dateFormat))
		suite.NoError(err, "No error when writing line")
	}
}

func (suite permitRepoSuite) TestWriteActivePermitsOfCarDuring_Positive() {
	const carId = "05539a50-6fac-c50d-b290-4e7372c573e9"
	const startDate = "2022-04-05"
	const endDate = "2022-04-16"
	permits, err := suite.permitRepo.GetActiveOfCarDuring(carId, startDate, endDate)
	suite.NoError(err, "No error when getting active permits of car during two timestamps")

	f, err := os.Create(fmt.Sprintf("testout/active_during_%s_%s.txt", startDate, endDate))
	suite.NoError(err, "No error creating file")
	defer f.Close()

	for _, permit := range permits {
		_, err := f.WriteString(permitToString(permit, suite.permitRepo.dateFormat))
		suite.NoError(err, "No error when writing line")
	}
}

func permitToString(permit models.Permit, dateFormat string) string {
	return fmt.Sprintf("%d,%s,%s,%s,%s,%d,%t\n",
		permit.Id,
		permit.ResidentId,
		permit.Car.Id,
		permit.StartDate.Format(dateFormat),
		permit.EndDate.Format(dateFormat),
		permit.RequestTS,
		permit.AffectsDays,
	)
}
