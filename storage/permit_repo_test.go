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

	migrator, err := GetUpMigrator(database)
	if err != nil {
		log.Fatal().Msgf("Failed to get migrator: %v", err)
	}
	suite.migrator = migrator

	suite.dateFormat = config.Constants().DateFormat()

	suite.existingCar = models.NewCar("fc377a4c-4a15-444d-85e7-ce8a3a578a8e", "OGYR3X", "blue", "", "", 6)
	suite.newPermit = models.NewNewPermitArgs("T1043321", suite.existingCar.ID,
		time.Now().Truncate(time.Second),
		time.Now().Add(time.Duration(24)*time.Hour).Truncate(time.Second),
		false,
		"")
}

func (suite permitRepoSuite) TestGetAllPermits_EmptySlice_Positive() {
	err := suite.migrator.Migrate(1)
	suite.NoError(err, "Error when migrating down to v1")
	defer func() {
		err := suite.migrator.Up()
		suite.NoError(err, "Error when migrating all the way up again")
	}()

	permits, err := suite.permitRepo.Get(models.AllPermits, defaultLimit, defaultOffset, false)
	suite.NoError(err, "Error getting all permits when the table is empty")
	suite.Equal(0, len(permits), "length of permits should be 0")
	suite.True(cmp.Equal(permits, []models.Permit{}), "permit should be an empty slice")
}

func (suite permitRepoSuite) TestGetActivePermits_EmptySlice_Positive() {
	err := suite.migrator.Migrate(1)
	suite.NoError(err, "Error when migrating down to v1")
	defer func() {
		err := suite.migrator.Up()
		suite.NoError(err, "Error when migrating all the way up again")
	}()

	permits, err := suite.permitRepo.Get(models.ActivePermits, defaultLimit, defaultOffset, false)
	suite.NoError(err, "Error getting active permits when the table is empty")
	suite.Equal(0, len(permits), "length of permits should be 0")
	suite.True(cmp.Equal(permits, []models.Permit{}), "permit should be an empty slice")
}

func (suite permitRepoSuite) TestGetAllPermits_NonEmpty_Positive() {
	permits, err := suite.permitRepo.Get(models.AllPermits, defaultLimit, defaultOffset, false)
	if err != nil {
		suite.NoError(err)
		return
	}
	suite.NotEqual(len(permits), 0, "length of permits should not be 0")
}

func (suite permitRepoSuite) TestGetAllPermits_Reversed_Positive() {
	permitID, _ := suite.permitRepo.Create(suite.newPermit)
	defer suite.permitRepo.Delete(permitID)

	permits, err := suite.permitRepo.Get(models.AllPermits, defaultLimit, defaultOffset, true)
	if err != nil {
		suite.NoError(err)
		return
	} else if len(permits) == 0 {
		suite.NotEqual(len(permits), 0, "no permits found")
		return
	}

	first := permits[0]
	suite.Equal(suite.newPermit.ResidentID, first.ResidentID)
	suite.Equal(suite.newPermit.CarID, first.Car.ID)
	suite.Empty(cmp.Diff(first.StartDate, suite.newPermit.StartDate))
	suite.Empty(cmp.Diff(first.EndDate, suite.newPermit.EndDate))
}

func (suite permitRepoSuite) TestGetExpiredPermits_NonEmpty_Positive() {
	_, err := suite.permitRepo.Get(models.ExpiredPermits, defaultLimit, defaultOffset, false)
	suite.NoError(err)
}

func (suite permitRepoSuite) TestGetCount_All_Positive() {
	_, err := suite.permitRepo.GetCount(models.AllPermits)
	suite.NoError(err)
}

func (suite permitRepoSuite) TestGetCount_Active_Positive() {
	_, err := suite.permitRepo.GetCount(models.ActivePermits)
	suite.NoError(err)
}

func (suite permitRepoSuite) TestGetCount_Expired_Positive() {
	_, err := suite.permitRepo.GetCount(models.ExpiredPermits)
	suite.NoError(err)
}

func (suite permitRepoSuite) TestGetCount_Exception_Positive() {
	_, err := suite.permitRepo.GetCount(models.ExceptionPermits)
	suite.NoError(err)
}

func (suite permitRepoSuite) TestWriteAllPermits_Positive() {
	permits, err := suite.permitRepo.Get(models.AllPermits, defaultLimit, defaultOffset, false)
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
	permits, err := suite.permitRepo.Get(models.ActivePermits, defaultLimit, defaultOffset, false)
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
	permits, err := suite.permitRepo.Get(models.ExceptionPermits, defaultLimit, defaultOffset, false)
	suite.NoError(err, "Error when getting permit exceptions")

	f, err := os.Create("testout/permit_exceptions.txt")
	suite.NoError(err, "Error creating permit_exceptions file")
	defer f.Close()

	for _, permit := range permits {
		_, err := f.WriteString(permitToString(permit, suite.dateFormat))
		suite.NoError(err, "Error when writing line")
	}
}

func (suite permitRepoSuite) TestGetOnePermit_Positive() {
	permitID, _ := suite.permitRepo.Create(suite.newPermit)
	defer suite.permitRepo.Delete(permitID)

	permit, err := suite.permitRepo.GetOne(permitID)
	suite.NoError(err)

	suite.Equal(permit.ResidentID, suite.newPermit.ResidentID)
	suite.Equal(permit.Car.ID, suite.newPermit.CarID)
	suite.Empty(cmp.Diff(permit.StartDate, suite.newPermit.StartDate))
	suite.Empty(cmp.Diff(permit.EndDate, suite.newPermit.EndDate))
}

func (suite permitRepoSuite) TestGetAllPermitsOfResident_Positive() {
	permitID, _ := suite.permitRepo.Create(suite.newPermit)
	defer suite.permitRepo.Delete(permitID)

	permits, err := suite.permitRepo.GetAllOfResident(suite.newPermit.ResidentID)
	if err != nil {
		suite.NoError(err)
		return
	} else if len(permits) == 0 {
		suite.NotEmpty(permits, "length of permits should not be zero")
		return
	}

	last := permits[len(permits)-1]
	suite.Equal(suite.newPermit.ResidentID, last.ResidentID)
	suite.Equal(suite.newPermit.CarID, last.Car.ID)
	suite.Empty(cmp.Diff(suite.newPermit.StartDate, last.StartDate))
	suite.Empty(cmp.Diff(suite.newPermit.EndDate, last.EndDate))
}

func (suite permitRepoSuite) TestGetActivePermitsOfResident_Positive() {
	permitID, _ := suite.permitRepo.Create(suite.newPermit)
	defer suite.permitRepo.Delete(permitID)

	permits, err := suite.permitRepo.GetActiveOfResident(suite.newPermit.ResidentID)
	if err != nil {
		suite.NoError(err)
		return
	} else if len(permits) == 0 {
		suite.NotEmpty(permits, "length of permits should not be zero")
		return
	}

	last := permits[len(permits)-1]
	suite.Equal(suite.newPermit.ResidentID, last.ResidentID)
	suite.Equal(suite.newPermit.CarID, last.Car.ID)
	suite.Empty(cmp.Diff(suite.newPermit.StartDate, last.StartDate))
	suite.Empty(cmp.Diff(suite.newPermit.EndDate, last.EndDate))
}

func (suite permitRepoSuite) TestGetActivePermitsOfCarDuring_StartBefore_EndBefore_Empty() {
	permitID, _ := suite.permitRepo.Create(suite.newPermit)
	defer suite.permitRepo.Delete(permitID)

	permits, err := func() ([]models.Permit, error) {
		startDate := suite.newPermit.StartDate.Add(time.Duration(-96) * time.Hour)
		endDate := suite.newPermit.StartDate.Add(time.Duration(-24) * time.Hour)
		return suite.permitRepo.GetActiveOfCarDuring(suite.existingCar.ID, startDate, endDate)
	}()

	suite.NoError(err, "Error when getting active permits of car during two timestamps")
	suite.Equal(0, len(permits), "length of permits should be 0")
}

func (suite permitRepoSuite) TestGetActivePermitsOfCarDuring_StartBefore_EndAtBeg_NonEmpty() {
	permitID, _ := suite.permitRepo.Create(suite.newPermit)
	defer suite.permitRepo.Delete(permitID)

	permits, err := func() ([]models.Permit, error) {
		startDate := suite.newPermit.StartDate.Add(time.Duration(-96) * time.Hour)
		endDate := suite.newPermit.StartDate
		return suite.permitRepo.GetActiveOfCarDuring(suite.existingCar.ID, startDate, endDate)
	}()

	suite.NoError(err, "Error when getting active permits of car during two timestamps")
	suite.Equal(1, len(permits), "length of permits should be 1")
}

func (suite permitRepoSuite) TestGetActivePermitsOfCarDuring_StartAtEnd_EndAfter_NonEmpty() {
	permitID, _ := suite.permitRepo.Create(suite.newPermit)
	defer suite.permitRepo.Delete(permitID)

	permits, err := func() ([]models.Permit, error) {
		startDate := suite.newPermit.EndDate
		endDate := suite.newPermit.EndDate.Add(time.Duration(96) * time.Hour)
		return suite.permitRepo.GetActiveOfCarDuring(suite.existingCar.ID, startDate, endDate)
	}()

	suite.NoError(err, "Error when getting active permits of car during two timestamps")
	suite.Equal(1, len(permits), "length of permits should be 1")
}

func (suite permitRepoSuite) TestGetActivePermitsOfCarDuring_StartAtBeg_EndAtEnd_NonEmpty() {
	permitID, _ := suite.permitRepo.Create(suite.newPermit)
	defer suite.permitRepo.Delete(permitID)

	permits, err := suite.permitRepo.GetActiveOfCarDuring(suite.existingCar.ID, suite.newPermit.StartDate, suite.newPermit.EndDate)
	suite.NoError(err, "Error when getting active permits of car during two timestamps")

	suite.Equal(1, len(permits), "length of permits should be 1")
}

func (suite permitRepoSuite) TestCreate_PermitDNE_Positive() {
	permitID, err := suite.permitRepo.Create(suite.newPermit)
	suite.NoError(err, "err from creating non-existing permit should be nil")

	suite.permitRepo.Delete(permitID)
}

func (suite permitRepoSuite) TestDelete_Positive() {
	permitID, _ := suite.permitRepo.Create(suite.newPermit)

	err := suite.permitRepo.Delete(permitID)
	suite.NoError(err, "err from deleting permit should be nil")
}

func (suite permitRepoSuite) TestSearch_PermitID_FullString_Positive() {
	permitID, _ := suite.permitRepo.Create(suite.newPermit)
	defer suite.permitRepo.Delete(permitID)

	permits, err := suite.permitRepo.Search(fmt.Sprint(permitID), models.AllPermits)
	if err != nil {
		suite.NoError(err)
		return
	} else if len(permits) == 0 {
		suite.NotEqual(len(permits), 0, "no search matches")
		return
	}

	foundPermit := containsID(permits, permitID)

	suite.True(foundPermit, "newly created permit was not among search matches")
}

func (suite permitRepoSuite) TestSearch_PermitID_SubString_Positive() {
	permitID, _ := suite.permitRepo.Create(suite.newPermit)
	defer suite.permitRepo.Delete(permitID)

	permits, err := suite.permitRepo.Search(fmt.Sprint(permitID)[1:], models.AllPermits)
	if err != nil {
		suite.NoError(err)
		return
	} else if len(permits) == 0 {
		suite.NotEqual(len(permits), 0, "no search matches")
		return
	}

	foundPermit := containsID(permits, permitID)

	suite.True(foundPermit, "newly created permit was not among search matches")
}

func (suite permitRepoSuite) TestSearch_LicensePlate_SubString_Positive() {
	permitID, _ := suite.permitRepo.Create(suite.newPermit)
	defer suite.permitRepo.Delete(permitID)

	permits, err := suite.permitRepo.Search(fmt.Sprint(suite.existingCar.LicensePlate)[1:], models.AllPermits)
	if err != nil {
		suite.NoError(err)
		return
	} else if len(permits) == 0 {
		suite.NotEqual(len(permits), 0, "no search matches")
		return
	}

	foundPermit := containsID(permits, permitID)

	suite.True(foundPermit, "newly created permit was not among search matches")
}

func permitToString(permit models.Permit, dateFormat string) string {
	return fmt.Sprintf("%d,%s,%s,%s,%s,%d,%t,%s\n",
		permit.ID,
		permit.ResidentID,
		permit.Car.ID,
		permit.StartDate.Format(dateFormat),
		permit.EndDate.Format(dateFormat),
		permit.RequestTS,
		permit.AffectsDays,
		permit.ExceptionReason,
	)
}

func containsID(permits []models.Permit, id int) bool {
	for _, permit := range permits {
		if permit.ID == id {
			return true
		}
	}
	return false
}
