BEGIN;

CREATE TABLE IF NOT EXISTS quota_adjustment(
  id SERIAL PRIMARY KEY UNIQUE NOT NULL,
  resident_id CHAR(8) REFERENCES resident(id) ON DELETE CASCADE NOT NULL,
  amount SMALLINT NOT NULL,
  reason TEXT NOT NULL,
  created_by_admin_id TEXT REFERENCES admin(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- one-time reconciliation: preserve every resident's currently-effective
-- limit as an explicit, auditable adjustment, so switching amt_parking_days_used
-- from a hand-editable counter to a derived value doesn't newly block anyone
INSERT INTO quota_adjustment (resident_id, amount, reason)
SELECT
  r.id,
  r.amt_parking_days_used - COALESCE(p.derived, 0),
  'Migration reconciliation: preserves previously-recorded allowance from legacy amt_parking_days_used field'
FROM resident r
LEFT JOIN (
  SELECT resident_id, SUM((end_ts - start_ts) / 86400) AS derived
  FROM permit
  WHERE affects_days = true
  GROUP BY resident_id
) p ON p.resident_id = r.id
WHERE r.amt_parking_days_used <> COALESCE(p.derived, 0);

COMMIT;
