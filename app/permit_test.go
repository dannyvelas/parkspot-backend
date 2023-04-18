package app

import (
	"context"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage/psql"
	"github.com/dannyvelas/lasvistas_api/util"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"net/http"
	"testing"
	"time"
)

type permitTestSuite struct {
	suite.Suite
	container       testcontainers.Container
	permitService   PermitService
	residentService ResidentService

	// the following are objects that will exist in db for duration of tests
	resident          models.Resident
	residentUnlimDays models.Resident
	car               models.Car

	// this map is shared between two tests so it is kept here
	desiredPermits map[string]models.Permit
}

func TestPermitRouter(t *testing.T) {
	suite.Run(t, new(permitTestSuite))
}

func (suite *permitTestSuite) SetupSuite() {
	// configure and start container
	container, database, err := getSandboxDatabase()
	if err != nil {
		suite.T().Fatalf("error getting sandbox database: %v", err)
	}
	// save container in suite struct so we can terminate it on suite teardown
	suite.container = container

	// define repos
	residentRepo := psql.NewResidentRepo(database)
	permitRepo := psql.NewPermitRepo(database)
	carRepo := psql.NewCarRepo(database)

	// define services
	suite.residentService = NewResidentService(residentRepo)
	carService := NewCarService(carRepo)
	suite.permitService = NewPermitService(permitRepo, residentRepo, carService)

	{ // create residents
		suite.resident = models.Resident{
			ID:        "B1234567",
			FirstName: "Daniel",
			LastName:  "Velasquez",
			Phone:     "1234567890",
			Email:     "email@example.com",
			Password:  "notapassword"}
		if _, err := suite.residentService.Create(suite.resident); err != nil {
			suite.TearDownSuite()
			suite.T().Fatalf("tearing down because failed to create resident: %v", err)
		}

		suite.residentUnlimDays = models.Resident{
			ID:        "B7654321",
			FirstName: "Daniel",
			LastName:  "Velasquez",
			Phone:     "1234567890",
			Email:     "email2@example.com",
			Password:  "notapassword",
			UnlimDays: util.ToPtr(true)}
		if _, err := suite.residentService.Create(suite.residentUnlimDays); err != nil {
			suite.TearDownSuite()
			suite.T().Fatalf("tearing down because failed to create resident: %v", err)
		}
	}

	{ // create car
		suite.car = models.NewCar(
			"d1e0affb-14e7-4e9f-b8a3-70be7d49d063",
			suite.resident.ID,
			"lp1",
			"color",
			"make",
			"model",
			0)
		_, err = carService.Create(suite.car)
		if err != nil {
			suite.TearDownSuite()
			suite.T().Fatalf("tearing down because failed to create resident: %v", err)
		}
	}

	suite.desiredPermits = map[string]models.Permit{
		"NoUnlimDays,NoException": addStartAndEndDate(models.Permit{ResidentID: suite.resident.ID, CarID: suite.car.ID}),
		"UnlimDays,NoException":   addStartAndEndDate(models.Permit{ResidentID: suite.residentUnlimDays.ID, CarID: suite.car.ID}),
		"NoUnlimDays,Exception":   addStartAndEndDate(models.Permit{ResidentID: suite.resident.ID, CarID: suite.car.ID, ExceptionReason: "some exception reason"}),
		"UnlimDays,Exception":     addStartAndEndDate(models.Permit{ResidentID: suite.residentUnlimDays.ID, CarID: suite.car.ID, ExceptionReason: "some exception reason"}),
	}
}

func (suite permitTestSuite) TearDownSuite() {
	err := suite.container.Terminate(context.Background())
	if err != nil {
		require.NoError(suite.T(), fmt.Errorf("error tearing down container: %v", err))
	}
}

func (suite permitTestSuite) TearDownTest() {
	err := suite.permitService.permitRepo.Reset()
	if err != nil {
		suite.T().Fatalf("encountered error resetting permit repo in-between tests")
	}
}

func (suite permitTestSuite) TestCreate_ResidentMultipleActivePermits() {
	_, err := suite.permitService.ValidateAndCreate(addStartAndEndDate(models.Permit{ResidentID: suite.resident.ID, CarID: suite.car.ID}))
	if err != nil {
		require.NoError(suite.T(), fmt.Errorf("Error creating permit before test: %v", err))
	}

	// initalize resident permits, each w a different car to eachother and to permitOne
	resPermitTwo := addStartAndEndDate(models.Permit{ResidentID: suite.resident.ID, LicensePlate: "two", Color: "two", Make: "two", Model: "two"})
	resPermitThree := addStartAndEndDate(models.Permit{ResidentID: suite.resident.ID, LicensePlate: "three", Color: "three", Make: "three", Model: "three"})

	type createPermitTest struct {
		name       string
		permit     models.Permit
		shouldBeOk bool
	}
	tests := []createPermitTest{
		{"resident second permit", resPermitTwo, true}, {"resident third permit", resPermitThree, false},
	}

	// see which permit creations succeed/fail
	for _, test := range tests {
		_, err := suite.permitService.ValidateAndCreate(test.permit)
		if err != nil && test.shouldBeOk {
			suite.NoError(fmt.Errorf("%s failed: Error creating permit when it should've been okay: %v", test.name, err))
		} else if err == nil && !test.shouldBeOk {
			suite.NoError(fmt.Errorf("%s failed: Permit was created when it shouldn't have", test.name))
		}
		// else: test passed: err == nil and should be okay OR err != nil and shouldn't be okay
	}
}

func (suite permitTestSuite) TestCreate_CarTwoActivePermits() {
	ogPermit, err := suite.permitService.ValidateAndCreate(addStartAndEndDate(models.Permit{ResidentID: suite.resident.ID, CarID: suite.car.ID}))
	if err != nil {
		require.NoError(suite.T(), fmt.Errorf("Error creating permit before test: %v", err))
	}

	permitSameCar := addStartAndEndDate(models.Permit{ResidentID: suite.resident.ID, CarID: ogPermit.CarID})

	_, err = suite.permitService.ValidateAndCreate(permitSameCar)
	require.ErrorIs(suite.T(), err, errs.CarActivePermit, "expected error to be car active permit")
}

func (suite permitTestSuite) TestCreate_NoStartNoEnd_ErrMissing() {
	desiredPermit := models.Permit{
		ResidentID:   suite.resident.ID,
		LicensePlate: "lp2",
		Color:        "color",
		Make:         "make",
		Model:        "model",
	}

	var apiErr *errs.ApiErr
	_, err := suite.permitService.ValidateAndCreate(desiredPermit)
	require.ErrorAs(suite.T(), err, &apiErr, "expected error to be instance of apiErr")

	suite.Equal(http.StatusBadRequest, apiErr.StatusCode)
}

func (suite permitTestSuite) TestCreate_AddsResDays() {
	for testName, desiredPermit := range suite.desiredPermits {
		residentBefore, err := suite.residentService.GetOne(desiredPermit.ResidentID)
		if err != nil {
			require.NoError(suite.T(), fmt.Errorf("%s failed: %v", testName, err))
		}

		createdPermit, err := suite.permitService.ValidateAndCreate(desiredPermit)
		if err != nil {
			require.NoError(suite.T(), fmt.Errorf("%s failed: %v", testName, err))
		}

		residentNow, err := suite.residentService.GetOne(desiredPermit.ResidentID)
		if err != nil {
			require.NoError(suite.T(), fmt.Errorf("%s failed: %v", testName, err))
		}

		shouldAddDays := desiredPermit.ResidentID == suite.resident.ID && desiredPermit.ExceptionReason == ""

		amtDaysAddedToRes := *residentNow.AmtParkingDaysUsed - *residentBefore.AmtParkingDaysUsed
		permitLength := util.GetAmtDays(desiredPermit.StartDate, desiredPermit.EndDate)
		if !shouldAddDays && amtDaysAddedToRes != 0 {
			err := fmt.Errorf("%s failed: added %d days when it shouldn't have. Permit length was: %d", testName, amtDaysAddedToRes, permitLength)
			require.NoError(suite.T(), err)
		} else if shouldAddDays && amtDaysAddedToRes != permitLength {
			err := fmt.Errorf("%s failed: added incorrect amount of days: %d instead of %d", testName, amtDaysAddedToRes, permitLength)
			require.NoError(suite.T(), err)
		}

		err = suite.permitService.Delete(createdPermit.ID)
		if err != nil {
			require.NoError(suite.T(), fmt.Errorf("%s failed: deleting test permit failed: %v", testName, err))
		}
	}
}

func (suite permitTestSuite) TestDelete_SubtractsResDays() {
	for testName, desiredPermit := range suite.desiredPermits {
		residentBefore, err := suite.residentService.GetOne(desiredPermit.ResidentID)
		if err != nil {
			require.NoError(suite.T(), fmt.Errorf("%s failed: %v", testName, err))
		}

		createdPermit, err := suite.permitService.ValidateAndCreate(desiredPermit)
		if err != nil {
			require.NoError(suite.T(), fmt.Errorf("%s failed: %v", testName, err))
		}

		err = suite.permitService.Delete(createdPermit.ID)
		if err != nil {
			require.NoError(suite.T(), fmt.Errorf("%s failed: %v", testName, err))
		}

		residentNow, err := suite.residentService.GetOne(desiredPermit.ResidentID)
		if err != nil {
			require.NoError(suite.T(), fmt.Errorf("%s failed: %v", testName, err))
		}

		if *residentBefore.AmtParkingDaysUsed != *residentNow.AmtParkingDaysUsed {
			require.NoError(suite.T(), fmt.Errorf("%s failed: did not subract days. Resident has %d instead of %d", testName, *residentNow.AmtParkingDaysUsed, *residentBefore.AmtParkingDaysUsed))
		}
	}
}

func (suite permitTestSuite) TestDelete_AddsCarDays() {
}

func (suite permitTestSuite) TestDelete_SubtractsCarDays() {
}

func (suite permitTestSuite) TestCreate_AllFieldsMatch() {
}

func (suite permitTestSuite) TestGetActivePermitsOfResident_Postive() {
	createdPermit, err := suite.permitService.ValidateAndCreate(addStartAndEndDate(models.Permit{ResidentID: suite.resident.ID, CarID: suite.car.ID}))
	if err != nil {
		require.NoError(suite.T(), fmt.Errorf("Error creating permit before test: %v", err))
	}

	permits, err := suite.permitService.GetAll(models.ActivePermits, config.MaxLimit, 0, true, "", suite.resident.ID)
	require.NoError(suite.T(), err)
	require.NotEmpty(suite.T(), permits.Records, "length of permits should not be zero")

	last := permits.Records[len(permits.Records)-1]

	require.Equal(suite.T(), createdPermit.ResidentID, last.ResidentID)
	require.Equal(suite.T(), createdPermit.LicensePlate, last.LicensePlate)
	require.Empty(suite.T(), cmp.Diff(createdPermit.StartDate, last.StartDate))
	require.Empty(suite.T(), cmp.Diff(createdPermit.EndDate, last.EndDate))
}

func (suite permitTestSuite) TestGetMaxExceptions_Positive() {
	createdPermit, err := suite.permitService.ValidateAndCreate(addStartAndEndDate(models.Permit{ResidentID: suite.resident.ID, CarID: suite.car.ID, ExceptionReason: "an exception reason here"}))
	if err != nil {
		require.NoError(suite.T(), fmt.Errorf("error creating permit before test: %v", err))
	}

	permits, err := suite.permitService.GetAll(models.ExceptionPermits, config.MaxLimit, 0, true, "", suite.resident.ID)
	require.NoError(suite.T(), err)
	require.NotEmpty(suite.T(), permits.Records, "length of permits should not be zero")

	if permits.Metadata.TotalAmount < config.MaxLimit {
		suite.Equal(permits.Metadata.TotalAmount, len(permits.Records), "The amount of records reported in metadata is lower than limit, so the amount of records in the payload should be equal to metadata.totalAmount")
	}

	last := permits.Records[len(permits.Records)-1]

	require.Equal(suite.T(), createdPermit.ResidentID, last.ResidentID)
	require.Equal(suite.T(), createdPermit.LicensePlate, last.LicensePlate)
	require.Empty(suite.T(), cmp.Diff(createdPermit.StartDate, last.StartDate))
	require.Empty(suite.T(), cmp.Diff(createdPermit.EndDate, last.EndDate))
}

// helpers
func addStartAndEndDate(permit models.Permit) models.Permit {
	permit.StartDate = time.Now().Truncate(time.Second)
	permit.EndDate = time.Now().Add(time.Hour * 24).Truncate(time.Second)
	return permit
}
