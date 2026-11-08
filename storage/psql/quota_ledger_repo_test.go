package psql

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dannyvelas/parkspot-backend/config"
	"github.com/dannyvelas/parkspot-backend/models"
	"github.com/dannyvelas/parkspot-backend/storage"
	"github.com/dannyvelas/parkspot-backend/util"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type quotaLedgerRepoTestSuite struct {
	suite.Suite
	container       testcontainers.Container
	quotaLedgerRepo storage.QuotaLedgerRepo
	permitRepo      storage.PermitRepo
}

func TestQuotaLedgerRepo(t *testing.T) {
	suite.Run(t, new(quotaLedgerRepoTestSuite))
}

func (suite *quotaLedgerRepoTestSuite) SetupSuite() {
	container, database, err := NewSandboxDatabase()
	if err != nil {
		suite.T().Fatalf("error getting sandbox database: %v", err)
	}
	suite.container = container
	suite.quotaLedgerRepo = database.QuotaLedgerRepo()
	suite.permitRepo = database.PermitRepo()

	if err := database.ResidentRepo().Create(models.TestResident); err != nil {
		suite.TearDownSuite()
		suite.T().Fatalf("tearing down because failed to create resident: %v", err)
	}
}

func (suite *quotaLedgerRepoTestSuite) TearDownSuite() {
	if err := suite.container.Terminate(context.Background()); err != nil {
		require.NoError(suite.T(), fmt.Errorf("error tearing down container: %v", err))
	}
}

func (suite *quotaLedgerRepoTestSuite) TearDownTest() {
	if err := suite.quotaLedgerRepo.Reset(); err != nil {
		suite.T().Fatalf("encountered error resetting quota ledger repo in-between tests: %v", err)
	}

	if err := suite.permitRepo.Reset(); err != nil {
		suite.T().Fatalf("encountered error resetting permit repo in-between tests: %v", err)
	}
}

func (suite *quotaLedgerRepoTestSuite) TestAdminAdjustmentRaisesEffectiveAllowance() {
	residentID := models.TestResident.ID
	quotaYearUTC := time.Now().UTC().Year()

	if _, err := suite.quotaLedgerRepo.CreatePermitEntry(permitLasting(residentID, 18, 0), 18); err != nil {
		require.NoError(suite.T(), fmt.Errorf("error creating the 18-day permit before test: %v", err))
	}

	balances, err := suite.quotaLedgerRepo.SelectBalances([]string{residentID}, quotaYearUTC)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(),
		models.QuotaBalance{DaysUsed: 18, EffectiveAllowance: config.MaxParkingDays},
		balances[residentID])

	threeDays := permitLasting(residentID, 3, time.Hour*24*20)
	_, err = suite.quotaLedgerRepo.CreatePermitEntry(threeDays, 3)
	require.Error(suite.T(), err, "the 3-day permit should be rejected at 18 of 20 days used")

	adjustment := models.NewQuotaLedger(
		residentID, quotaYearUTC, 5, models.EntryTypeAdminAdjustment,
		nil, nil, util.ToPtr("board-approved extension"),
	)
	if _, err := suite.quotaLedgerRepo.CreateAdjustment(adjustment); err != nil {
		require.NoError(suite.T(), fmt.Errorf("error recording the admin adjustment: %v", err))
	}

	balances, err = suite.quotaLedgerRepo.SelectBalances([]string{residentID}, quotaYearUTC)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(),
		models.QuotaBalance{DaysUsed: 18, EffectiveAllowance: config.MaxParkingDays + 5},
		balances[residentID])

	if _, err := suite.quotaLedgerRepo.CreatePermitEntry(threeDays, 3); err != nil {
		require.NoError(suite.T(), fmt.Errorf("the 3-day permit should be allowed after a +5 adjustment: %v", err))
	}

	balances, err = suite.quotaLedgerRepo.SelectBalances([]string{residentID}, quotaYearUTC)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(),
		models.QuotaBalance{DaysUsed: 21, EffectiveAllowance: config.MaxParkingDays + 5},
		balances[residentID])
}

func (suite *quotaLedgerRepoTestSuite) TestCancelPermitEntryTwiceCreditsTheDaysOnce() {
	residentID := models.TestResident.ID
	quotaYearUTC := time.Now().UTC().Year()

	createdPermit, err := suite.quotaLedgerRepo.CreatePermitEntry(permitLasting(residentID, 10, 0), 10)
	require.NoError(suite.T(), err)

	balances, err := suite.quotaLedgerRepo.SelectBalances([]string{residentID}, quotaYearUTC)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 10, balances[residentID].DaysUsed)

	require.NoError(suite.T(), suite.quotaLedgerRepo.CancelPermitEntry(createdPermit, 10))

	balances, err = suite.quotaLedgerRepo.SelectBalances([]string{residentID}, quotaYearUTC)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 0, balances[residentID].DaysUsed)

	require.NoError(suite.T(), suite.quotaLedgerRepo.CancelPermitEntry(createdPermit, 10),
		"a repeated cancellation should succeed silently rather than error")

	balances, err = suite.quotaLedgerRepo.SelectBalances([]string{residentID}, quotaYearUTC)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 0, balances[residentID].DaysUsed,
		"a second cancellation must not credit the days back again")

	history, err := suite.quotaLedgerRepo.SelectHistory(residentID)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), history, 2,
		"expected exactly one creation entry and one cancellation entry")
}

func permitLasting(residentID string, days int, offset time.Duration) models.Permit {
	startDate := time.Now().Add(offset).Truncate(time.Second)

	return models.Permit{
		ResidentID:  residentID,
		CarID:       models.TestCar.ID,
		StartDate:   startDate,
		EndDate:     startDate.Add(time.Hour * 24 * time.Duration(days)),
		AffectsDays: true,
	}
}
