package storage

import (
	"errors"
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
	carRepo                carRepo
	migrator               *migrate.Migrate
	existingCar            models.Car
	existingCarEmptyFields models.Car
	nonExistingCar         models.Car
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
	suite.carRepo = newCarRepo(database)

	migrator, err := GetV1Migrator(database)
	if err != nil {
		log.Fatal().Msgf("Failed to get migrator: %v", err)
	}
	suite.migrator = migrator

	if err := suite.migrator.Up(); err != nil {
		log.Fatal().Msgf("Error when migrating all the way up: %v", err)
	}

	suite.existingCar = models.NewCar("8976e334-e281-7efd-ae84-92171d53434b", "VHS1K3A", "orange", "BMW", "X3")
	suite.existingCarEmptyFields = models.NewCar("fc377a4c-4a15-544d-c5e7-ce8a3a578a8e", "OGYR3X", "blue", "", "")
	suite.nonExistingCar = models.NewCar("1dc45c1b-e686-4668-a07b-fc49086408cf", "ABC123", "red", "toyota", "tercel")
}

func (suite carRepoSuite) TearDownSuite() {
	err := suite.migrator.Down()
	if err != nil {
		suite.NoError(err, "No error migrating all the way down")
	}
}

func (suite carRepoSuite) TestGetOne_Negative() {
	_, err := suite.carRepo.GetOne("549c3b81-f3ca-49a3-8a79-a472c7f4554a") // non-existent uuid
	suite.ErrorIs(err, ErrNoRows, "err should be equal to storage.ErrNoRows")
}

func (suite carRepoSuite) TestGetOne_NULLFields_Positive() {
	existingCarEmptyFields := suite.existingCarEmptyFields

	foundCar, err := suite.carRepo.GetOne(existingCarEmptyFields.Id)
	suite.NoError(err, "no errors when getting one car with empty fields")

	// check that they're equal. not using `suite.Equal` because it doesn't let you define your own Equal() func
	suite.Empty(cmp.Diff(foundCar, existingCarEmptyFields), "car should be equal to testCar")
}

func (suite carRepoSuite) TestGetOne_NoNULLFields_Positive() {
	existingCar := suite.existingCar

	foundCar, err := suite.carRepo.GetOne(existingCar.Id)
	suite.NoError(err, "no errors when getting one car without any empty fields")

	// check that they're equal. not using `suite.Equal` because it doesn't let you define your own Equal() func
	suite.Empty(cmp.Diff(foundCar, existingCar), "car should be equal to testCar")
}

func (suite carRepoSuite) TestCreate_EmptyFields_Negative() {
	for emptyFieldName, inCar := range genEmptyFields(suite.nonExistingCar) {
		outCar, err := suite.carRepo.Create(inCar)
		condition := errors.Is(err, models.ErrInvalidFields) || errors.Is(err, models.ErrEmptyFields)
		suite.Truef(condition, "err should be models.ErrInValidFields or models.ErrEmptyFields for car without %s. Was: %v", emptyFieldName, err)
		suite.Empty(cmp.Diff(outCar, models.Car{}), "car should be equal to Car{}")
	}
}

func (suite carRepoSuite) TestCreate_CarExists_Negative() {
	car, err := suite.carRepo.Create(suite.existingCar)
	suite.NotNil(err, "err from creating existingCar should not be nil")
	suite.Empty(cmp.Diff(car, models.Car{}), "car should be equal to Car{}")
}

func (suite carRepoSuite) TestCreateIfNotExists_EmptyFields_Negative() {
	for emptyFieldName, inCar := range genEmptyFields(suite.nonExistingCar) {
		outCar, err := suite.carRepo.CreateIfNotExists(inCar)
		if emptyFieldName == "Id" {
			suite.ErrorIsf(err, ErrEmptyIDArg, "err should wrap storage.ErrEmptyIDArg for car without %s. Instead was: %v", emptyFieldName, err)
		} else {
			condition := errors.Is(err, models.ErrInvalidFields) || errors.Is(err, models.ErrEmptyFields)
			suite.Truef(condition, "err should be models.ErrInValidFields or models.ErrEmptyFields for car without %s. Was: %v", emptyFieldName, err)
		}
		suite.Empty(cmp.Diff(outCar, models.Car{}), "car returned should be equal to Car{}")
	}
}

func (suite carRepoSuite) TestCreateIfNotExists_CarExists_Positive() {
	foundCar, err := suite.carRepo.CreateIfNotExists(suite.existingCar)
	suite.Nil(err, "err from creating existingCar should be nil")
	suite.Empty(cmp.Diff(foundCar, suite.existingCar), "car found should be equal to the existingCar passed in")
}
