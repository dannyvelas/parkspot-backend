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
	existingCarEmptyFields models.Car
	newCar                 models.NewCarArgs
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

	migrator, err := GetUpMigrator(database)
	if err != nil {
		log.Fatal().Msgf("Failed to get migrator: %v", err)
	}
	suite.migrator = migrator

	suite.existingCarEmptyFields = models.NewCar("fc377a4c-4a15-444d-85e7-ce8a3a578a8e", "OGYR3X", "blue", "", "", 6)
	suite.newCar = models.NewNewCarArgs("ABC123", "red", "toyota", "tercel")
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
	carId, _ := suite.carRepo.Create(suite.newCar)
	defer suite.carRepo.Delete(carId)
	existingCar := suite.newCar.ToCar(carId)

	foundCar, err := suite.carRepo.GetOne(carId)
	suite.NoError(err, "Error when getting one car without any empty fields")

	// check that they're equal. not using `suite.Equal` because it doesn't let you define your own Equal() func
	suite.Empty(cmp.Diff(foundCar, existingCar), "car should be equal to testCar")
}

func (suite carRepoSuite) TestGetByLicensePlate_Positive() {
	carId, _ := suite.carRepo.Create(suite.newCar)
	defer suite.carRepo.Delete(carId)
	existingCar := suite.newCar.ToCar(carId)

	foundCar, err := suite.carRepo.GetByLicensePlate(existingCar.LicensePlate)
	suite.NoError(err, "Error when getting one car by its license plate")

	// check that they're equal. not using `suite.Equal` because it doesn't let you define your own Equal() func
	suite.Empty(cmp.Diff(foundCar, existingCar), "car should be equal to testCar")
}

func (suite carRepoSuite) TestGetByLicensePlate_Negative() {
	_, err := suite.carRepo.GetByLicensePlate("ABC123") // non-existent license plate
	suite.ErrorIs(err, ErrNoRows, "err should be equal to storage.ErrNoRows")
}

func (suite carRepoSuite) TestCreate_CarExists_Negative() {
	carId, _ := suite.carRepo.Create(suite.newCar)
	defer suite.carRepo.Delete(carId)

	_, err := suite.carRepo.Create(suite.newCar)
	suite.Error(err, "err from creating existing car should not be nil")
}

func (suite carRepoSuite) TestCreate_CarDNE_Positive() {
	carId, err := suite.carRepo.Create(suite.newCar)
	suite.NoError(err, "err from creating non-existing car should not be nil")

	suite.carRepo.Delete(carId)
}

func (suite carRepoSuite) TestDelete_Positive() {
	carId, _ := suite.carRepo.Create(suite.newCar)

	err := suite.carRepo.Delete(carId)
	suite.NoError(err, "err from deleting car should be nil")
}
