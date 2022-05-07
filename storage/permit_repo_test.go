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
	location    *time.Location
	permitRepo  PermitRepo
	migrator    *migrate.Migrate
	dateFormat  string
	existingCar models.Car
	newPermit   models.NewPermitArgs
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
	suite.existingCar = models.NewCar("fc377a4c-4a15-544d-c5e7-ce8a3a578a8e", "OGYR3X", "blue", "", "", 6)
	suite.newPermit = models.NewNewPermitArgs("T1043321", "fc377a4c-4a15-544d-c5e7-ce8a3a578a8e",
		time.Date(2022, 06, 18, 0, 0, 0, 0, time.Local),
		time.Date(2022, 06, 29, 0, 0, 0, 0, time.Local),
		1645279579,
		false,
		nil)
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
	suite.NoError(err, "Error getting all permits when the table is empty")
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
	permits, err := suite.permitRepo.GetAll(defaultLimit, defaultOffset)
	suite.NoError(err, "Error getting all permits when the table is not empty")
	suite.NotEqual(len(permits), 0, "length of permits should not be 0")
}

func (suite permitRepoSuite) TestGetActivePermits_NonEmpty_Positive() {
	permits, err := suite.permitRepo.GetActive(defaultLimit, defaultOffset)
	suite.NoError(err, "Error getting active permits when the table is not empty")
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

func (suite permitRepoSuite) TestWritePermitExceptions_Positive() {
	permits, err := suite.permitRepo.GetExceptions(defaultLimit, defaultOffset)
	suite.NoError(err, "Error when getting permit exceptions")

	f, err := os.Create("testout/permit_exceptions.txt")
	suite.NoError(err, "Error creating permit_exceptions file")
	defer f.Close()

	for _, permit := range permits {
		_, err := f.WriteString(permitToString(permit, suite.dateFormat))
		suite.NoError(err, "Error when writing line")
	}
}

func (suite permitRepoSuite) TestGetActivePermitsOfCarDuring_StartBefore_EndBefore_Empty() {
	permitId, _ := suite.permitRepo.Create(suite.newPermit)
	defer suite.permitRepo.Delete(permitId)

	permits, err := func() ([]models.Permit, error) {
		startDate := suite.newPermit.StartDate.Add(time.Duration(-96) * time.Hour)
		endDate := suite.newPermit.StartDate.Add(time.Duration(-24) * time.Hour)
		return suite.permitRepo.GetActiveOfCarDuring(suite.existingCar.Id, startDate, endDate)
	}()

	suite.NoError(err, "Error when getting active permits of car during two timestamps")
	suite.Equal(0, len(permits), "length of permits should be 0")
}

func (suite permitRepoSuite) TestGetActivePermitsOfCarDuring_StartBefore_EndAtBeg_NonEmpty() {
	permitId, _ := suite.permitRepo.Create(suite.newPermit)
	defer suite.permitRepo.Delete(permitId)

	permits, err := func() ([]models.Permit, error) {
		startDate := suite.newPermit.StartDate.Add(time.Duration(-96) * time.Hour)
		endDate := suite.newPermit.StartDate
		return suite.permitRepo.GetActiveOfCarDuring(suite.existingCar.Id, startDate, endDate)
	}()

	suite.NoError(err, "Error when getting active permits of car during two timestamps")
	suite.Equal(1, len(permits), "length of permit should be 1")
}

func (suite permitRepoSuite) TestGetActivePermitsOfCarDuring_StartAtEnd_EndAfter_NonEmpty() {
	permitId, _ := suite.permitRepo.Create(suite.newPermit)
	defer suite.permitRepo.Delete(permitId)

	permits, err := func() ([]models.Permit, error) {
		startDate := suite.newPermit.EndDate
		endDate := suite.newPermit.EndDate.Add(time.Duration(96) * time.Hour)
		return suite.permitRepo.GetActiveOfCarDuring(suite.existingCar.Id, startDate, endDate)
	}()

	suite.NoError(err, "Error when getting active permits of car during two timestamps")
	suite.Equal(1, len(permits), "length of permit should be 1")
}

func (suite permitRepoSuite) TestGetActivePermitsOfCarDuring_StartAtBeg_EndAtEnd_NonEmpty() {
	permitId, _ := suite.permitRepo.Create(suite.newPermit)
	defer suite.permitRepo.Delete(permitId)

	permits, err := suite.permitRepo.GetActiveOfCarDuring(suite.existingCar.Id, suite.newPermit.StartDate, suite.newPermit.EndDate)
	suite.NoError(err, "Error when getting active permits of car during two timestamps")

	suite.Equal(1, len(permits), "length of permit should be 1")
}

func (suite permitRepoSuite) TestCreate_PermitDNE_Positive() {
	permitId, err := suite.permitRepo.Create(suite.newPermit)
	suite.NoError(err, "err from creating non-existing permit should be nil")

	suite.permitRepo.Delete(permitId)
}

func (suite permitRepoSuite) TestDelete_Positive() {
	permitId, _ := suite.permitRepo.Create(suite.newPermit)

	err := suite.permitRepo.Delete(permitId)
	suite.NoError(err, "err from deleting permit should be nil")
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
