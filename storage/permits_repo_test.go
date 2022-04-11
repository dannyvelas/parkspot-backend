package storage

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/golang-migrate/migrate/v4"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
	"reflect"
	"testing"
)

type permitsRepoSuite struct {
	suite.Suite
	permitsRepo PermitsRepo
	migrator    *migrate.Migrate
}

func TestPermitsRepo(t *testing.T) {
	suite.Run(t, new(permitsRepoSuite))
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
	suite.NoError(err, "no error getting all permits when the table is empty")
	suite.Equal(len(permits), 0, "length of permits should be 0")
	suite.Equal(permits, []models.Permit{}, "permits should be an empty slice")
}

func (suite permitsRepoSuite) TestGetActivePermits_EmptySlice_Positive() {
	permits, err := suite.permitsRepo.GetActive(defaultLimit, defaultOffset)
	suite.NoError(err, "no error getting active permits when the table is empty")
	suite.Equal(len(permits), 0, "length of permits should be 0")
	suite.Equal(permits, []models.Permit{}, "permits should be an empty slice")
}

func (suite permitsRepoSuite) TestGetAllPermits_NonEmpty_Positive() {
	err := suite.migrator.Up()
	suite.NoError(err, "no error when migrating all the way up")

	permits, err := suite.permitsRepo.GetAll(defaultLimit, defaultOffset)
	suite.NoError(err, "no error getting all permits when the table is not empty")
	suite.NotEqual(len(permits), 0, "length of permits should not be 0")

	for _, permit := range permits {
		suite.NoError(allFieldsNonZero(permit))
		suite.NoError(allFieldsNonZero(permit.Car))
	}
}

func allFieldsNonZero(aStruct any) error {
	structV := reflect.ValueOf(aStruct)
	structT := reflect.TypeOf(aStruct)

	for i, field := range reflect.VisibleFields(structT) {
		if field.Type.Kind() != reflect.Bool && structV.Field(i).IsZero() {
			return fmt.Errorf("%s.%s should not be the zero value for its type", structT.Name(), field.Name)
		}
	}
	return nil
}
