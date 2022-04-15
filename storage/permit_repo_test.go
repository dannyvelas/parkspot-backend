package storage

import (
	//"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/golang-migrate/migrate/v4"
	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
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

	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Fatal().Msgf("Failed to load location: %v", err)
		return
	}
	suite.location = location

	database, err := NewDatabase(config.Postgres())
	if err != nil {
		log.Fatal().Msgf("Failed to start database: %v", err)
		return
	}

	migrator, err := GetMigrator(database)
	if err != nil {
		log.Fatal().Msgf("Failed to get migrator: %v", err)
	}

	suite.permitRepo = NewPermitRepo(database)
	suite.migrator = migrator
}

func (suite permitRepoSuite) TearDownTest() {
	suite.permitRepo.deleteAll()
}

func (suite permitRepoSuite) TearDownSuite() {
	suite.migrator.Down()
}

func (suite permitRepoSuite) TestGetAllPermits_EmptySlice_Positive() {
	permits, err := suite.permitRepo.GetAll(defaultLimit, defaultOffset)
	suite.NoError(err, "no error getting all permit when the table is empty")
	suite.Equal(len(permits), 0, "length of permit should be 0")
	suite.Equal(permits, []models.Permit{}, "permit should be an empty slice")
}

func (suite permitRepoSuite) TestGetActivePermits_EmptySlice_Positive() {
	permits, err := suite.permitRepo.GetActive(defaultLimit, defaultOffset)
	suite.NoError(err, "no error getting active permits when the table is empty")
	suite.Equal(len(permits), 0, "length of permits should be 0")
	suite.Equal(permits, []models.Permit{}, "permits should be an empty slice")
}

func (suite permitRepoSuite) TestGetAllPermits_NonEmpty_Positive() {
	err := suite.migrator.Up()
	suite.NoError(err, "no error when migrating all the way up")

	// check that length is not 0
	permits, err := suite.permitRepo.GetAll(defaultLimit, defaultOffset)
	suite.NoError(err, "no error getting all permits when the table is not empty")
	suite.NotEqual(len(permits), 0, "length of permits should not be 0")

	// create dates
	startDate, err := time.ParseInLocation("2006-01-02", "2022-02-22", suite.location)
	suite.NoError(err, "no error creating startDate")
	endDate, err := time.ParseInLocation("2006-01-02", "2022-03-05", suite.location)
	suite.NoError(err, "no error creating endDate")

	// create test permit using above dates
	testCar := models.Car{Id: "fc377a4c-4a15-544d-c5e7-ce8a3a578a8e", LicensePlate: "OGYR3X",
		Color: "blue", Make: "", Model: ""}
	testPermit := models.Permit{Id: 1, ResidentId: "T1043321", Car: testCar, StartDate: startDate,
		EndDate: endDate, RequestTS: 1645487283, AffectsDays: true}

	// get first permit
	firstPermit := permits[0]

	// check that they're equal. not using `suite.Equal` because it doesn't let you define your own Equal() func
	suite.Empty(cmp.Diff(firstPermit, testPermit), "firstPermit should be equal to testPermit")
}
