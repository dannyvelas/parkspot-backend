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
	existingCarEmptyFields models.Car
	existingCreateCar      models.CreateCar
	nonExistingCreateCar   models.CreateCar
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

	suite.existingCar = models.NewCar("9b3080d5-6fb7-8271-504d-281fc9535b63", "HYMQC1A7", "red", "SKI-DOO", "SKANDIC WT E-TEC 600 HO", 23)
	suite.existingCarEmptyFields = models.NewCar("fc377a4c-4a15-544d-c5e7-ce8a3a578a8e", "OGYR3X", "blue", "", "", 6)
	suite.existingCreateCar = models.NewCreateCar("HYMQC1A7", "red", "SKI-DOO", "SKANDIC WT E-TEC 600 HO")
	suite.nonExistingCreateCar = models.NewCreateCar("ABC123", "red", "toyota", "tercel")
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

func (suite carRepoSuite) TestGetByLicensePlate_Positive() {
	existingCar := suite.existingCar

	foundCar, err := suite.carRepo.GetByLicensePlate(existingCar.LicensePlate)
	suite.NoError(err, "Error when getting one car by its license plate")

	// check that they're equal. not using `suite.Equal` because it doesn't let you define your own Equal() func
	suite.Empty(cmp.Diff(foundCar, existingCar), "car should be equal to testCar")
}

func (suite carRepoSuite) TestGetByLicensePlate_Negative() {
	_, err := suite.carRepo.GetByLicensePlate("ABCD123") // non-existent license plate
	suite.ErrorIs(err, ErrNoRows, "err should be equal to storage.ErrNoRows")
}

func (suite carRepoSuite) TestCreate_CarExists_Negative() {
	_, err := suite.carRepo.Create(suite.existingCreateCar)
	suite.Errorf(err, "err from creating existing car %v should not be nil", suite.existingCar)
}

func (suite carRepoSuite) TestCreate_CarDNE_Positive() {
	newCar, err := suite.carRepo.Create(suite.nonExistingCreateCar)
	suite.NoErrorf(err, "err from creating non-existing car %v should not be nil")
	suite.Empty(cmp.Diff(newCar, suite.nonExistingCreateCar.ToCar(newCar.Id)), "newCar should be equal to nonExistingCreateCar")
}
