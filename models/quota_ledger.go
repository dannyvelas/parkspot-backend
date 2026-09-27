package models

import (
	"time"
)

type EntryType string

const (
	EntryTypePermit          EntryType = "PERMIT"
	EntryTypeAdminAdjustment EntryType = "ADMIN_ADJUSTMENT"
)

type QuotaLedger struct {
	ID               int       `json:"id"`
	ResidentID       string    `json:"residentID"`
	QuotaYearUTC     int       `json:"quotaYearUTC"`
	Amount           int       `json:"amount"`
	EntryType        EntryType `json:"entryType"`
	PermitID         *int      `json:"permitID"`
	Note             *string   `json:"reason"`
	CreatedByAdminID *string   `json:"createdByAdminID"`
	CreatedAt        time.Time `json:"createdAt"`
}

func NewQuotaLedger(
	residentID string,
	quotaYearUTC int,
	amount int,
	entryType EntryType,
	permitID *int,
	createdByAdminID *string,
	note *string,
) QuotaLedger {
	return QuotaLedger{
		ResidentID:       residentID,
		QuotaYearUTC:     quotaYearUTC,
		Amount:           amount,
		EntryType:        entryType,
		PermitID:         permitID,
		CreatedByAdminID: createdByAdminID,
		Note:             note,
	}
}

// QuotaBalance is a resident's derived usage and effective allowance for a
// single allowance period, computed from QuotaLedgerRepo.SelectBalances.
type QuotaBalance struct {
	DaysUsed           int
	EffectiveAllowance int
}
