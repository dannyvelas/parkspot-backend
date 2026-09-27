BEGIN;

CREATE TABLE IF NOT EXISTS quota_ledger(
  id SERIAL PRIMARY KEY UNIQUE NOT NULL,
  resident_id CHAR(8) REFERENCES resident(id) ON DELETE CASCADE NOT NULL,
  quota_year_utc INT NOT NULL,
  amount SMALLINT NOT NULL,
  entry_type TEXT NOT NULL CHECK (entry_type IN ('PERMIT', 'ADMIN_ADJUSTMENT')),
  permit_id INT REFERENCES permit(id),
  created_by_admin_id TEXT REFERENCES admin(id),
  note TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS quota_ledger_resident_year_utc_idx ON quota_ledger (resident_id, quota_year_utc);

ALTER TABLE permit ADD COLUMN IF NOT EXISTS cancelled_ts BIGINT;

-- backfill: one PERMIT-type ledger entry per existing qualifying permit, preserving
-- its original creation time and the allowance period (quota_year_utc) that was in
-- effect then, so pre-cutover activity is exactly as individually auditable as
-- anything created after this migration (see specs/002-unify-quota-ledger/spec.md FR-013)
INSERT INTO quota_ledger (resident_id, quota_year_utc, amount, entry_type, permit_id, created_at)
SELECT
  p.resident_id,
  EXTRACT(YEAR FROM (to_timestamp(p.request_ts) AT TIME ZONE 'UTC'))::int,
  (p.end_ts - p.start_ts) / 86400,
  'PERMIT',
  p.id,
  to_timestamp(p.request_ts)
FROM permit p
WHERE p.affects_days = true;

-- reconciliation: close any residual gap between a resident's legacy running-total
-- counter and what was just backfilled above, so switching amt_parking_days_used
-- from a hand-editable counter to a derived value doesn't newly block anyone
INSERT INTO quota_ledger (resident_id, quota_year_utc, amount, entry_type, note)
SELECT
  r.id,
  EXTRACT(YEAR FROM (now() AT TIME ZONE 'UTC'))::int,
  r.amt_parking_days_used - COALESCE(b.total, 0),
  'ADMIN_ADJUSTMENT',
  'Migration reconciliation: preserves previously-recorded allowance from legacy amt_parking_days_used field'
FROM resident r
LEFT JOIN (
  SELECT resident_id, SUM(amount) AS total
  FROM quota_ledger
  WHERE entry_type = 'PERMIT'
  GROUP BY resident_id
) b ON b.resident_id = r.id
WHERE r.amt_parking_days_used <> COALESCE(b.total, 0);

COMMIT;
