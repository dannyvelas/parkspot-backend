package storage

import (
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/golang-migrate/migrate/v4"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
	"testing"
)

type permitsRepoSuite struct {
	suite.Suite
	permitsRepo PermitsRepo
	migrator    *migrate.Migrate
}

func (suite *permitsRepoSuite) SetupSuite() {
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

	suite.permitsRepo = NewPermitsRepo(database)
	suite.migrator = migrator
}

func (suite permitsRepoSuite) TearDownTest() {
	suite.permitsRepo.deleteAll()
}

func (suite permitsRepoSuite) TearDownSuite() {
	suite.migrator.Down()
}

func (suite permitsRepoSuite) TestGetAllPermits_EmptySlice_Positive() {
	permits, err := suite.permitsRepo.GetAll(defaultLimit, defaultOffset)
	suite.NoError(err, "no error when getting all permits when the table is empty")
	suite.Equal(len(permits), 0, "length of permits should be 0")
	suite.Equal(permits, []models.Permit{}, "permits should be an empty slice")
}

func TestPermitsRepo(t *testing.T) {
	suite.Run(t, new(permitsRepoSuite))
}
