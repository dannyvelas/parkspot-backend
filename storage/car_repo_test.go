package storage

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/golang-migrate/migrate/v4"
	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
	"reflect"
	"testing"
	"time"
)

type carRepoSuite struct {
	suite.Suite
	location *time.Location
	carRepo  carRepo
	migrator *migrate.Migrate
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
	car, err := suite.carRepo.GetOne("fc377a4c-4a15-544d-c5e7-ce8a3a578a8e")
	suite.NoError(err, "no errors when getting one car")

	testCar := models.NewCar("fc377a4c-4a15-544d-c5e7-ce8a3a578a8e", "OGYR3X", "blue", "", "")

	// check that they're equal. not using `suite.Equal` because it doesn't let you define your own Equal() func
	suite.Empty(cmp.Diff(car, testCar), "car should be equal to testCar")
}

func (suite carRepoSuite) TestGetOne_NoNULLFields_Positive() {
	car, err := suite.carRepo.GetOne("8976e334-e281-7efd-ae84-92171d53434b")
	suite.NoError(err, "no errors when getting one car")

	testCar := models.NewCar("8976e334-e281-7efd-ae84-92171d53434b", "VHS1K3A", "orange", "BMW", "X3")

	// check that they're equal. not using `suite.Equal` because it doesn't let you define your own Equal() func
	suite.Empty(cmp.Diff(car, testCar), "car should be equal to testCar")
}

func (suite carRepoSuite) TestCreate_EmptyFields_Negative() {
	carAllFields := models.NewCar("1dc45c1b-e686-4668-a07b-fc49086408cf", "ABC123", "red", "toyota", "tercel")
	for _, field := range reflect.VisibleFields(reflect.TypeOf(carAllFields)) {
		// make each field "", one by one
		carMissingField := carAllFields
		carMissingFieldV := reflect.ValueOf(&carMissingField).Elem()
		fieldV := carMissingFieldV.FieldByName(field.Name)
		previousVal := fieldV.Interface()
		fieldV.SetString("")

		car, err := suite.carRepo.Create(carMissingField)
		suite.ErrorIs(err, ErrMissingField, "errors.Is(err, ErrMissingField) should be true")
		suite.Equal(err.Error(), fmt.Sprintf("%s: [%s]", ErrMissingField.message, field.Name))
		suite.Empty(cmp.Diff(car, models.Car{}), "car should be equal to Car{}")

		fieldV.SetString(previousVal.(string)) // restore value
	}
}

func (suite carRepoSuite) TestCreate_CarExists_Negative() {
	// chose car with no empty fields, otherwise this would fail for another reason - empty fields not allowed
	existingCar := models.NewCar("8976e334-e281-7efd-ae84-92171d53434b", "VHS1K3A", "orange", "BMW", "X3")

	car, err := suite.carRepo.Create(existingCar)
	suite.NotNil(err, "err from creating existingCar should not be nil")
	suite.Empty(cmp.Diff(car, models.Car{}), "car should be equal to Car{}")
}
