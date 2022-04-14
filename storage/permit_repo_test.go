package storage

import (
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type permitRepoSuite struct {
	suite.Suite
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
	suite.Equal(permits, []Permit{}, "permit should be an empty slice")
}

func (suite permitRepoSuite) TestGetActivePermits_EmptySlice_Positive() {
	permits, err := suite.permitRepo.GetActive(defaultLimit, defaultOffset)
	suite.NoError(err, "no error getting active permits when the table is empty")
	suite.Equal(len(permits), 0, "length of permits should be 0")
	suite.Equal(permits, []Permit{}, "permits should be an empty slice")
}

func (suite permitRepoSuite) TestGetAllPermits_NonEmpty_Positive() {
	err := suite.migrator.Up()
	suite.NoError(err, "no error when migrating all the way up")

	permits, err := suite.permitRepo.GetAll(defaultLimit, defaultOffset)
	suite.NoError(err, "no error getting all permits when the table is not empty")
	suite.NotEqual(len(permits), 0, "length of permits should not be 0")

	startDate, err := time.Parse("2006-01-02", "2022-02-08")
	suite.NoError(err, "no error creating startDate")

	endDate, err := time.Parse("2006-01-02", "2022-02-08")
	suite.NoError(err, "no error creating startDate")

	testCar := Car{
		Id: "7f8186e8-0070-462c-bc07-39a6f29f0f6a", LicensePlate: "HVELOMM",
		Color: "green", Make: "BMW", Model: "X3"}
	testPermit := Permit{
		Id: 1, ResidentId: "B1580553", Car: testCar, StartDate: startDate,
		EndDate: endDate, RequestTS: 1644213010, AffectsDays: false}

	firstPermit := permits[0]

	suite.Equal(firstPermit, testPermit, "first permit should be equal to test permit")

	//for _, permit := range permits {
	//	suite.NoError(allFieldsNonZero(permit))
	//	suite.NoError(allFieldsNonZero(permit.Car))
	//}
}
