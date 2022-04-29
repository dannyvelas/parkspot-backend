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
	dateFormat string
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
	suite.permitRepo = NewPermitRepo(database)

	migrator, err := GetV1Migrator(database)
	if err != nil {
		log.Fatal().Msgf("Failed to get migrator: %v", err)
	}
	suite.migrator = migrator

	if err := suite.migrator.Up(); err != nil {
		log.Fatal().Msgf("Error when migrating all the way up: %v", err)
	}

	suite.dateFormat = "2006-01-02"
}

func (suite permitRepoSuite) TearDownSuite() {
	err := suite.migrator.Down()
	if err != nil {
		suite.NoError(err, "Error migrating all the way down")
	}
}

func (suite permitRepoSuite) TestGetAllPermits_EmptySlice_Positive() {
	err := suite.migrator.Migrate(1)
	suite.NoError(err, "Error when migrating down to v1")
	defer func() {
		err := suite.migrator.Up()
		suite.NoError(err, "Error when migrating all the way up again")
	}()

	permits, err := suite.permitRepo.GetAll(defaultLimit, defaultOffset)
	suite.NoError(err, "Error getting all permit when the table is empty")
	suite.Equal(0, len(permits), "length of permit should be 0")
	suite.True(cmp.Equal(permits, []models.Permit{}), "permit should be an empty slice")
}

func (suite permitRepoSuite) TestGetActivePermits_EmptySlice_Positive() {
	err := suite.migrator.Migrate(1)
	suite.NoError(err, "Error when migrating down to v1")
	defer func() {
		err := suite.migrator.Up()
		suite.NoError(err, "Error when migrating all the way up again")
	}()

	permits, err := suite.permitRepo.GetActive(defaultLimit, defaultOffset)
	suite.NoError(err, "Error getting active permits when the table is empty")
	suite.Equal(0, len(permits), "length of permits should be 0")
	suite.True(cmp.Equal(permits, []models.Permit{}), "permit should be an empty slice")
}

func (suite permitRepoSuite) TestGetAllPermits_NonEmpty_Positive() {
	// check that length is not 0
	permits, err := suite.permitRepo.GetAll(defaultLimit, defaultOffset)
	suite.NoError(err, "Error getting all permits when the table is not empty")
	suite.NotEqual(len(permits), 0, "length of permits should not be 0")
}

func (suite permitRepoSuite) TestGetActivePermits_NonEmpty_Positive() {
	// check that length is not 0
	permits, err := suite.permitRepo.GetActive(defaultLimit, defaultOffset)
	suite.NoError(err, "Error getting all permits when the table is not empty")
	suite.NotEqual(len(permits), 0, "length of permits should not be 0")
}

func (suite permitRepoSuite) TestWriteAllPermits_Positive() {
	permits, err := suite.permitRepo.GetAll(defaultLimit, defaultOffset)
	suite.NoError(err, "Error when getting all permits")

	f, err := os.Create("testout/all_permits.txt")
	suite.NoError(err, "Error creating all_permits file")
	defer f.Close()

	for _, permit := range permits {
		_, err := f.WriteString(permitToString(permit, suite.dateFormat))
		suite.NoError(err, "Error when writing line")
	}
}

func (suite permitRepoSuite) TestWriteActivePermits_Positive() {
	permits, err := suite.permitRepo.GetActive(defaultLimit, defaultOffset)
	suite.NoError(err, "Error when getting active permits")

	f, err := os.Create("testout/active_permits.txt")
	suite.NoError(err, "Error creating active_permits file")
	defer f.Close()

	for _, permit := range permits {
		_, err := f.WriteString(permitToString(permit, suite.dateFormat))
		suite.NoError(err, "Error when writing line")
	}
}

func (suite permitRepoSuite) TestWriteActivePermitsOfCarDuring_Positive() {
	const carId = "05539a50-6fac-c50d-b290-4e7372c573e9"
	startDate, err := time.ParseInLocation(suite.dateFormat, "2022-04-05", time.Local)
	suite.NoError(err, "Error parsing start date")

	endDate, err := time.ParseInLocation(suite.dateFormat, "2022-04-16", time.Local)
	suite.NoError(err, "Error parsing end date")

	permits, err := suite.permitRepo.GetActiveOfCarDuring(carId, startDate, endDate)
	suite.NoError(err, "Error when getting active permits of car during two timestamps")

	f, err := os.Create(fmt.Sprintf("testout/active_during_%s_%s.txt", startDate, endDate))
	suite.NoError(err, "Error creating file")
	defer f.Close()

	for _, permit := range permits {
		_, err := f.WriteString(permitToString(permit, suite.dateFormat))
		suite.NoError(err, "Error when writing line")
	}
}

func (suite permitRepoSuite) TestCreate_PermitDNE_Positive() {
	nonExistingCreatePermit := func() models.CreatePermit {
		existingResidentId := "T1043321"
		startDate := time.Date(2022, 06, 18, 0, 0, 0, 0, time.Local)
		endDate := time.Date(2022, 06, 29, 0, 0, 0, 0, time.Local)
		return models.NewCreatePermit(existingResidentId, models.CreateCar{}, startDate, endDate, time.Now().Unix(), false, nil)
	}()

	existingCarId := "fc377a4c-4a15-544d-c5e7-ce8a3a578a8e"
	_, err := suite.permitRepo.Create(nonExistingCreatePermit, existingCarId)
	suite.NoError(err, "err from creating non-existing permit should be nil")
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
