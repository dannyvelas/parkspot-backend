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

type carRepoSuite struct {
	suite.Suite
	carRepo                CarRepo
	migrator               *migrate.Migrate
	existingCar            models.Car
	existingCreateCar      models.CreateCar
	existingCarEmptyFields models.Car
}

func TestCarRepo(t *testing.T) {
	suite.Run(t, new(carRepoSuite))
}

func (suite *carRepoSuite) SetupSuite() {
	config := config.NewConfig()

	database, err := NewDatabase(config.Postgres())
	if err != nil {
		log.Fatal().Msgf("Failed to start database: %v", err)
	}
	suite.carRepo = NewCarRepo(database)

	migrator, err := GetV1Migrator(database)
	if err != nil {
		log.Fatal().Msgf("Failed to get migrator: %v", err)
	}
	suite.migrator = migrator

	if err := suite.migrator.Up(); err != nil {
		log.Fatal().Msgf("Error when migrating all the way up: %v", err)
	}

	suite.existingCar = models.NewCar("8976e334-e281-7efd-ae84-92171d53434b", "VHS1K3A", "orange", "BMW", "X3")
	suite.existingCreateCar = models.NewCreateCar("VHS1K3A", "orange", "BMW", "X3")
	suite.existingCarEmptyFields = models.NewCar("fc377a4c-4a15-544d-c5e7-ce8a3a578a8e", "OGYR3X", "blue", "", "")
}

func (suite carRepoSuite) TearDownSuite() {
	err := suite.migrator.Down()
	if err != nil {
		suite.NoError(err, "Error migrating all the way down")
	}
}

func (suite carRepoSuite) TestGetOne_Negative() {
	_, err := suite.carRepo.GetOne("549c3b81-f3ca-49a3-8a79-a472c7f4554a") // non-existent uuid
	suite.ErrorIs(err, ErrNoRows, "err should be equal to storage.ErrNoRows")
}

func (suite carRepoSuite) TestGetOne_NULLFields_Positive() {
	existingCarEmptyFields := suite.existingCarEmptyFields

	foundCar, err := suite.carRepo.GetOne(existingCarEmptyFields.Id)
	suite.NoError(err, "Error when getting one car with empty fields")

	// check that they're equal. not using `suite.Equal` because it doesn't let you define your own Equal() func
	suite.Empty(cmp.Diff(foundCar, existingCarEmptyFields), "car should be equal to testCar")
}

func (suite carRepoSuite) TestGetOne_NoNULLFields_Positive() {
	existingCar := suite.existingCar

	foundCar, err := suite.carRepo.GetOne(existingCar.Id)
	suite.NoError(err, "Error when getting one car without any empty fields")

	// check that they're equal. not using `suite.Equal` because it doesn't let you define your own Equal() func
	suite.Empty(cmp.Diff(foundCar, existingCar), "car should be equal to testCar")
}

func (suite carRepoSuite) TestCreate_CarExists_Negative() {
	car, err := suite.carRepo.Create(suite.existingCreateCar)
	suite.NotNil(err, "err from creating existingCar should not be nil")
	suite.Empty(cmp.Diff(car, models.Car{}), "car should be equal to Car{}")
}
