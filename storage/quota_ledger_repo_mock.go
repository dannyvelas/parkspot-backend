package storage

import (
	"time"

	"github.com/dannyvelas/parkspot-backend/config"
	"github.com/dannyvelas/parkspot-backend/errs"
	"github.com/dannyvelas/parkspot-backend/models"
)

type QuotaLedgerRepoMock struct {
	entries      []models.QuotaLedger
	nextEntryID  int
	nextPermitID int
}

func NewQuotaLedgerRepoMock() QuotaLedgerRepoMock {
	return QuotaLedgerRepoMock{}
}

// balance replicates QuotaLedgerRepo's SQL aggregate in plain Go: no locking
// is simulated here (mocks can't meaningfully simulate row-level locking);
// the concurrency guarantee is verified against the real database instead.
func (quotaLedgerRepoMock *QuotaLedgerRepoMock) balance(residentID string, quotaYearUTC int) (daysUsed int, adjustmentSum int) {
	for _, entry := range quotaLedgerRepoMock.entries {
		if entry.ResidentID != residentID || entry.QuotaYearUTC != quotaYearUTC {
			continue
		}
		daysUsed += entry.Amount
		if entry.EntryType == models.EntryTypeAdminAdjustment {
			adjustmentSum += entry.Amount
		}
	}
	return daysUsed, adjustmentSum
}

func (quotaLedgerRepoMock *QuotaLedgerRepoMock) CreatePermitEntry(desiredPermit models.Permit, permitLength int) (models.Permit, error) {
	quotaLedgerRepoMock.nextPermitID++
	createdPermit := desiredPermit
	createdPermit.ID = quotaLedgerRepoMock.nextPermitID

	if !desiredPermit.AffectsDays {
		return createdPermit, nil
	}

	quotaYearUTC := time.Now().UTC().Year()
	daysUsed, adjustmentSum := quotaLedgerRepoMock.balance(desiredPermit.ResidentID, quotaYearUTC)
	effectiveAllowance := config.MaxParkingDays + adjustmentSum

	if daysUsed >= effectiveAllowance {
		return models.Permit{}, errs.EntityDaysTooLong("resident", daysUsed)
	} else if daysUsed+permitLength > effectiveAllowance {
		return models.Permit{}, errs.PermitPlusEntityDaysTooLong("resident", daysUsed)
	}

	permitID := createdPermit.ID
	quotaLedgerRepoMock.nextEntryID++
	quotaLedgerRepoMock.entries = append(quotaLedgerRepoMock.entries, models.QuotaLedger{
		ID:           quotaLedgerRepoMock.nextEntryID,
		ResidentID:   desiredPermit.ResidentID,
		QuotaYearUTC: quotaYearUTC,
		Amount:       permitLength,
		EntryType:    models.EntryTypePermit,
		PermitID:     &permitID,
		CreatedAt:    time.Now(),
	})

	return createdPermit, nil
}

func (quotaLedgerRepoMock *QuotaLedgerRepoMock) CancelPermitEntry(permit models.Permit, permitLength int) error {
	if !permit.AffectsDays {
		return nil
	}

	// same formula as creation (FR-011): the permit's own original
	// request_ts, never the current time (FR-017).
	quotaYearUTC := time.Unix(permit.RequestTS, 0).UTC().Year()
	permitID := permit.ID
	quotaLedgerRepoMock.nextEntryID++
	quotaLedgerRepoMock.entries = append(quotaLedgerRepoMock.entries, models.QuotaLedger{
		ID:           quotaLedgerRepoMock.nextEntryID,
		ResidentID:   permit.ResidentID,
		QuotaYearUTC: quotaYearUTC,
		Amount:       -permitLength,
		EntryType:    models.EntryTypePermit,
		PermitID:     &permitID,
		CreatedAt:    time.Now(),
	})

	return nil
}

func (quotaLedgerRepoMock *QuotaLedgerRepoMock) CreateAdjustment(entry models.QuotaLedger) (models.QuotaLedger, error) {
	quotaLedgerRepoMock.nextEntryID++
	entry.ID = quotaLedgerRepoMock.nextEntryID
	entry.CreatedAt = time.Now()
	quotaLedgerRepoMock.entries = append(quotaLedgerRepoMock.entries, entry)

	return entry, nil
}

func (quotaLedgerRepoMock *QuotaLedgerRepoMock) SelectHistory(residentID string) ([]models.QuotaLedger, error) {
	history := []models.QuotaLedger{}
	for _, entry := range quotaLedgerRepoMock.entries {
		if entry.ResidentID == residentID {
			history = append(history, entry)
		}
	}

	return history, nil
}

func (quotaLedgerRepoMock *QuotaLedgerRepoMock) SelectBalances(residentIDs []string, quotaYearUTC int) (map[string]models.QuotaBalance, error) {
	wanted := make(map[string]bool, len(residentIDs))
	for _, id := range residentIDs {
		wanted[id] = true
	}

	daysUsed := make(map[string]int)
	adjustmentSum := make(map[string]int)
	for _, entry := range quotaLedgerRepoMock.entries {
		if entry.QuotaYearUTC != quotaYearUTC || !wanted[entry.ResidentID] {
			continue
		}
		daysUsed[entry.ResidentID] += entry.Amount
		if entry.EntryType == models.EntryTypeAdminAdjustment {
			adjustmentSum[entry.ResidentID] += entry.Amount
		}
	}

	balances := make(map[string]models.QuotaBalance, len(daysUsed))
	for residentID, used := range daysUsed {
		balances[residentID] = models.QuotaBalance{
			DaysUsed:           used,
			EffectiveAllowance: config.MaxParkingDays + adjustmentSum[residentID],
		}
	}

	return balances, nil
}

func (quotaLedgerRepoMock *QuotaLedgerRepoMock) Reset() error {
	quotaLedgerRepoMock.entries = quotaLedgerRepoMock.entries[:0]
	quotaLedgerRepoMock.nextEntryID = 0
	quotaLedgerRepoMock.nextPermitID = 0
	return nil
}
