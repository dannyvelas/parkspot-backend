package storage

import (
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/golang-migrate/migrate/v4"
	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
	"testing"
)

type residentRepoSuite struct {
	suite.Suite
	residentRepo     ResidentRepo
	migrator         *migrate.Migrate
	existingResident models.Resident
}

func TestResidentRepo(t *testing.T) {
	suite.Run(t, new(residentRepoSuite))
}

func (suite *residentRepoSuite) SetupSuite() {
	config := config.NewConfig()

	database, err := NewDatabase(config.Postgres())
	if err != nil {
		log.Fatal().Msgf("Failed to start database: %v", err)
	}
	suite.residentRepo = NewResidentRepo(database)

	migrator, err := GetUpMigrator(database)
	if err != nil {
		log.Fatal().Msgf("Failed to get migrator: %v", err)
	}
	suite.migrator = migrator

	suite.existingResident = models.NewResident("T1043321", "John", "Gibson", "(894) 280-4660", "john.gibson@gmail.com", "5730ec12ad69b442d69319417dce5869", true, 13)
}

func (suite residentRepoSuite) TestGetOne_Negative() {
	_, err := suite.residentRepo.GetOne("549c3b81-f3ca-49a3-8a79-a472c7f4554a") // non-existent uuid
	suite.ErrorIs(err, ErrNoRows, "err should be equal to storage.ErrNoRows")
}

func (suite residentRepoSuite) TestGetOne_Positive() {
	existingResident := suite.existingResident

	foundResident, err := suite.residentRepo.GetOne(existingResident.Id)
	suite.NoError(err, "Error when getting one existing resident")

	// check that they're equal. not using `suite.Equal` because it doesn't let you define your own Equal() func
	suite.Empty(cmp.Diff(foundResident, existingResident), "resident found should be equal to existing resident")
}
