package psql

import (
	"database/sql"
	"time"

	"github.com/dannyvelas/parkspot-backend/models"
)

type quotaLedger struct {
	ID               int            `db:"id"`
	ResidentID       string         `db:"resident_id"`
	QuotaYearUTC     int            `db:"quota_year_utc"`
	Amount           int            `db:"amount"`
	EntryType        string         `db:"entry_type"` // scanned as plain text; converted to models.EntryType in toModels()
	PermitID         sql.NullInt64  `db:"permit_id"`
	CreatedByAdminID sql.NullString `db:"created_by_admin_id"`
	Note             sql.NullString `db:"note"`
	CreatedAt        time.Time      `db:"created_at"`
}

func (q quotaLedger) toModels() models.QuotaLedger {
	var permitID *int
	if q.PermitID.Valid {
		id := int(q.PermitID.Int64)
		permitID = &id
	}

	var createdByAdminID *string
	if q.CreatedByAdminID.Valid {
		createdByAdminID = &q.CreatedByAdminID.String
	}

	var note *string
	if q.Note.Valid {
		note = &q.Note.String
	}

	return models.QuotaLedger{
		ID:               q.ID,
		ResidentID:       q.ResidentID,
		QuotaYearUTC:     q.QuotaYearUTC,
		Amount:           q.Amount,
		EntryType:        models.EntryType(q.EntryType),
		PermitID:         permitID,
		CreatedByAdminID: createdByAdminID,
		Note:             note,
		CreatedAt:        q.CreatedAt,
	}
}

type quotaLedgerSlice []quotaLedger

func (entries quotaLedgerSlice) toModels() []models.QuotaLedger {
	modelsEntries := make([]models.QuotaLedger, 0, len(entries))
	for _, entry := range entries {
		modelsEntries = append(modelsEntries, entry.toModels())
	}
	return modelsEntries
}
