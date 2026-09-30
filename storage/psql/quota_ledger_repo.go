package psql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/parkspot-backend/config"
	"github.com/dannyvelas/parkspot-backend/errs"
	"github.com/dannyvelas/parkspot-backend/models"
	"github.com/dannyvelas/parkspot-backend/storage"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type QuotaLedgerRepo struct {
	driver *sqlx.DB
}

func NewQuotaLedgerRepo(driver *sqlx.DB) storage.QuotaLedgerRepo {
	return QuotaLedgerRepo{driver: driver}
}

// sqlGetter is satisfied by both *sqlx.DB and *sqlx.Tx, so insertPermit can
// run outside a transaction (when a permit doesn't affect days) or inside
// one (when it does and its ledger entry must be atomic with it).
type sqlGetter interface {
	Get(dest interface{}, query string, args ...interface{}) error
}

func (r QuotaLedgerRepo) CreatePermitEntry(desiredPermit models.Permit, permitLength int) (models.Permit, error) {
	if !desiredPermit.AffectsDays {
		return r.insertPermit(r.driver, desiredPermit, time.Now().Unix())
	}

	tx, err := r.driver.Beginx()
	if err != nil {
		return models.Permit{}, fmt.Errorf("quota_ledger_repo.CreatePermitEntry: %w: %v", errs.ErrDBExec, err)
	}
	defer tx.Rollback()

	// FOR UPDATE tells Postgres to lock this resident row exclusively for the
	// rest of this transaction, effectively using it as a mutex: if two
	// CreatePermitEntry calls for the same resident run concurrently, the
	// second one blocks here until the first commits or rolls back.
	if _, err := tx.Exec("SELECT 1 FROM resident WHERE id = $1 FOR UPDATE", desiredPermit.ResidentID); err != nil {
		return models.Permit{}, fmt.Errorf("quota_ledger_repo.CreatePermitEntry: %w: %v", errs.ErrDBQuery, err)
	}

	now := time.Now()
	quotaYearUTC := now.UTC().Year()

	var daysUsed, adjustmentSum int
	err = tx.QueryRow(`
    SELECT COALESCE(SUM(amount), 0),
           COALESCE(SUM(amount) FILTER (WHERE entry_type = $1), 0)
    FROM quota_ledger
    WHERE resident_id = $2 AND quota_year_utc = $3
  `, string(models.EntryTypeAdminAdjustment), desiredPermit.ResidentID, quotaYearUTC).Scan(&daysUsed, &adjustmentSum)
	if err != nil {
		return models.Permit{}, fmt.Errorf("quota_ledger_repo.CreatePermitEntry: %w: %v", errs.ErrDBQuery, err)
	}

	effectiveAllowance := config.MaxParkingDays + adjustmentSum
	if daysUsed >= effectiveAllowance {
		return models.Permit{}, errs.EntityDaysTooLong("resident", daysUsed)
	} else if daysUsed+permitLength > effectiveAllowance {
		return models.Permit{}, errs.PermitPlusEntityDaysTooLong("resident", daysUsed)
	}

	createdPermit, err := r.insertPermit(tx, desiredPermit, now.Unix())
	if err != nil {
		return models.Permit{}, err
	}

	_, err = tx.Exec(`
    INSERT INTO quota_ledger (resident_id, quota_year_utc, amount, entry_type, permit_id)
    VALUES ($1, $2, $3, $4, $5)
  `, desiredPermit.ResidentID, quotaYearUTC, permitLength, string(models.EntryTypePermit), createdPermit.ID)
	if err != nil {
		return models.Permit{}, fmt.Errorf("quota_ledger_repo.CreatePermitEntry: %w: %v", errs.ErrDBExec, err)
	}

	if err := tx.Commit(); err != nil {
		return models.Permit{}, fmt.Errorf("quota_ledger_repo.CreatePermitEntry: %w: %v", errs.ErrDBExec, err)
	}

	return createdPermit, nil
}

// insertPermit issues the same INSERT permitRepo.Create does, but returns the
// full created permit in one round trip (via RETURNING) instead of a
// create-then-GetOne pair, and accepts either the plain driver or an
// in-flight transaction so CreatePermitEntry can keep it atomic with the
// ledger entry it inserts alongside it.
func (r QuotaLedgerRepo) insertPermit(g sqlGetter, desiredPermit models.Permit, requestTS int64) (models.Permit, error) {
	nullableReason := sql.NullString{}
	if desiredPermit.ExceptionReason != "" {
		nullableReason = sql.NullString{String: desiredPermit.ExceptionReason, Valid: true}
	}

	query, args, err := stmtBuilder.
		Insert("permit").
		SetMap(squirrel.Eq{
			"resident_id":      desiredPermit.ResidentID,
			"car_id":           desiredPermit.CarID,
			"license_plate":    desiredPermit.LicensePlate,
			"color":            desiredPermit.Color,
			"make":             desiredPermit.Make,
			"model":            desiredPermit.Model,
			"start_ts":         desiredPermit.StartDate.Unix(),
			"end_ts":           desiredPermit.EndDate.Unix(),
			"request_ts":       requestTS,
			"affects_days":     desiredPermit.AffectsDays,
			"exception_reason": nullableReason,
		}).
		Suffix("RETURNING id AS permit_id, resident_id, car_id, license_plate, color, make, model, start_ts, end_ts, request_ts, affects_days, exception_reason").
		ToSql()
	if err != nil {
		return models.Permit{}, fmt.Errorf("quota_ledger_repo.insertPermit: %w: %v", errs.ErrDBBuildingQuery, err)
	}

	row := permit{}
	if err := g.Get(&row, query, args...); err != nil {
		return models.Permit{}, fmt.Errorf("quota_ledger_repo.insertPermit: %w: %v", errs.ErrDBExec, err)
	}

	return row.toModels(), nil
}

func (r QuotaLedgerRepo) CancelPermitEntry(permit models.Permit, permitLength int) error {
	now := time.Now().Unix()

	if !permit.AffectsDays {
		if _, err := r.driver.Exec("UPDATE permit SET cancelled_ts = $1 WHERE id = $2", now, permit.ID); err != nil {
			return fmt.Errorf("quota_ledger_repo.CancelPermitEntry: %w: %v", errs.ErrDBExec, err)
		}
		return nil
	}

	tx, err := r.driver.Beginx()
	if err != nil {
		return fmt.Errorf("quota_ledger_repo.CancelPermitEntry: %w: %v", errs.ErrDBExec, err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE permit SET cancelled_ts = $1 WHERE id = $2", now, permit.ID); err != nil {
		return fmt.Errorf("quota_ledger_repo.CancelPermitEntry: %w: %v", errs.ErrDBExec, err)
	}

	// same quota_year_utc formula as creation (FR-011), computed from the
	// permit's own original request_ts, never from the current time — a
	// cancellation must net against the period it was originally booked
	// against, not whichever period happens to be current now (FR-017).
	quotaYearUTC := time.Unix(permit.RequestTS, 0).UTC().Year()
	if _, err := tx.Exec(`
    INSERT INTO quota_ledger (resident_id, quota_year_utc, amount, entry_type, permit_id)
    VALUES ($1, $2, $3, $4, $5)
  `, permit.ResidentID, quotaYearUTC, -permitLength, string(models.EntryTypePermit), permit.ID); err != nil {
		return fmt.Errorf("quota_ledger_repo.CancelPermitEntry: %w: %v", errs.ErrDBExec, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("quota_ledger_repo.CancelPermitEntry: %w: %v", errs.ErrDBExec, err)
	}

	return nil
}

func (r QuotaLedgerRepo) CreateAdjustment(entry models.QuotaLedger) (models.QuotaLedger, error) {
	tx, err := r.driver.Beginx()
	if err != nil {
		return models.QuotaLedger{}, fmt.Errorf("quota_ledger_repo.CreateAdjustment: %w: %v", errs.ErrDBExec, err)
	}
	defer tx.Rollback()

	// locks the same resident row CreatePermitEntry locks, so an adjustment
	// can't race a concurrent permit-creation check for the same resident.
	if _, err := tx.Exec("SELECT 1 FROM resident WHERE id = $1 FOR UPDATE", entry.ResidentID); err != nil {
		return models.QuotaLedger{}, fmt.Errorf("quota_ledger_repo.CreateAdjustment: %w: %v", errs.ErrDBQuery, err)
	}

	nullableAdmin := sql.NullString{}
	if entry.CreatedByAdminID != nil {
		nullableAdmin = sql.NullString{String: *entry.CreatedByAdminID, Valid: true}
	}
	nullableNote := sql.NullString{}
	if entry.Note != nil {
		nullableNote = sql.NullString{String: *entry.Note, Valid: true}
	}

	row := quotaLedger{}
	err = tx.Get(&row, `
    INSERT INTO quota_ledger (resident_id, quota_year_utc, amount, entry_type, created_by_admin_id, note)
    VALUES ($1, $2, $3, $4, $5, $6)
    RETURNING *
  `, entry.ResidentID, entry.QuotaYearUTC, entry.Amount, string(entry.EntryType), nullableAdmin, nullableNote)
	if err != nil {
		return models.QuotaLedger{}, fmt.Errorf("quota_ledger_repo.CreateAdjustment: %w: %v", errs.ErrDBExec, err)
	}

	if err := tx.Commit(); err != nil {
		return models.QuotaLedger{}, fmt.Errorf("quota_ledger_repo.CreateAdjustment: %w: %v", errs.ErrDBExec, err)
	}

	return row.toModels(), nil
}

func (r QuotaLedgerRepo) SelectHistory(residentID string) ([]models.QuotaLedger, error) {
	rows := quotaLedgerSlice{}
	err := r.driver.Select(&rows, "SELECT * FROM quota_ledger WHERE resident_id = $1 ORDER BY created_at ASC", residentID)
	if err != nil {
		return nil, fmt.Errorf("quota_ledger_repo.SelectHistory: %w: %v", errs.ErrDBQuery, err)
	}

	return rows.toModels(), nil
}

func (r QuotaLedgerRepo) SelectBalances(residentIDs []string, quotaYearUTC int) (map[string]models.QuotaBalance, error) {
	type balanceRow struct {
		ResidentID    string `db:"resident_id"`
		DaysUsed      int    `db:"days_used"`
		AdjustmentSum int    `db:"adjustment_sum"`
	}

	rows := []balanceRow{}
	err := r.driver.Select(&rows, `
    SELECT resident_id,
           COALESCE(SUM(amount), 0) AS days_used,
           COALESCE(SUM(amount) FILTER (WHERE entry_type = $1), 0) AS adjustment_sum
    FROM quota_ledger
    WHERE resident_id = ANY($2) AND quota_year_utc = $3
    GROUP BY resident_id
  `, string(models.EntryTypeAdminAdjustment), pq.Array(residentIDs), quotaYearUTC)
	if err != nil {
		return nil, fmt.Errorf("quota_ledger_repo.SelectBalances: %w: %v", errs.ErrDBQuery, err)
	}

	balances := make(map[string]models.QuotaBalance, len(rows))
	for _, row := range rows {
		balances[row.ResidentID] = models.QuotaBalance{
			DaysUsed:           row.DaysUsed,
			EffectiveAllowance: config.MaxParkingDays + row.AdjustmentSum,
		}
	}

	return balances, nil
}

func (r QuotaLedgerRepo) Reset() error {
	_, err := r.driver.Exec("DELETE FROM quota_ledger")
	if err != nil {
		return fmt.Errorf("quota_ledger_repo.Reset: %w: %v", errs.ErrDBExec, err)
	}

	return nil
}
