package storage

import (
	"github.com/dannyvelas/parkspot-backend/models"
)

// QuotaLedgerRepo is the single source of truth for every event that changes
// a resident's parking-day usage, whether generated automatically alongside a
// permit's creation/cancellation or recorded manually by an admin. Entries
// are permanent once created: this interface deliberately exposes no
// Update/Delete method, so it is structurally impossible for a caller to
// mutate or remove one (spec FR-002).
type QuotaLedgerRepo interface {
	// CreatePermitEntry inserts a permit and, if it affects days, a
	// same-transaction PERMIT-type ledger entry for permitLength days,
	// atomically checking the resident's effective allowance for the
	// permit's quota year first.
	CreatePermitEntry(permit models.Permit, permitLength int) (models.Permit, error)
	// CancelPermitEntry soft-cancels a permit and, if it affected days,
	// inserts a same-transaction offsetting PERMIT-type ledger entry
	// against the same quota year as the permit's original entry.
	CancelPermitEntry(permit models.Permit, permitLength int) error
	// CreateAdjustment records one manual (ADMIN_ADJUSTMENT) ledger entry.
	CreateAdjustment(entry models.QuotaLedger) (models.QuotaLedger, error)
	// SelectHistory returns a resident's complete ledger history in
	// chronological order.
	SelectHistory(residentID string) ([]models.QuotaLedger, error)
	// SelectBalances computes each of the given residents' derived usage
	// and effective allowance for quotaYearUTC in a single query. A resident
	// with no ledger rows for that year is simply absent from the
	// returned map; callers must default that case to
	// models.QuotaBalance{DaysUsed: 0, EffectiveAllowance: config.MaxParkingDays}.
	SelectBalances(residentIDs []string, quotaYearUTC int) (map[string]models.QuotaBalance, error)
	Reset() error // for testing purposes
}
