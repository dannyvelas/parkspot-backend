# Phase 1 Data Model: Separate Permit Usage from Admin-Editable Allowance

## Entities

### Resident (modified)

Existing table `resident`. Field-level change only — no new columns; one column removed.

| Field | Type | Notes |
|---|---|---|
| `id` | `CHAR(8)` PK | unchanged |
| `first_name`, `last_name`, `phone`, `email`, `password` | unchanged | unchanged |
| `unlim_days` | `BOOLEAN NOT NULL DEFAULT FALSE` | unchanged — still the sole bypass flag for the day-limit check |
| ~~`amt_parking_days_used`~~ | ~~`SMALLINT`~~ | **REMOVED** (migration `000008`). Replaced by two derived, non-persisted values computed at read time: |
| *derived* `days_used` | `int` (not a column) | `SUM((permit.end_ts - permit.start_ts) / 86400)` over `permit WHERE resident_id = resident.id AND affects_days = true`. Never settable via any write path. |
| *derived* `effective_allowance` | `int` (not a column) | `config.MaxParkingDays (20) + SUM(quota_adjustment.amount WHERE resident_id = resident.id)` |
| `token_version` | unchanged | unchanged |

**Validation rules**:
- A permit create request is rejected when `days_used + requested_permit_days > effective_allowance`, unless `unlim_days = true` or the permit is an exception (`exception_reason != ""`). *(FR-006, unchanged bypass conditions from current behavior.)*
- No API request body field may set `days_used` or `effective_allowance` directly; they are response-only computed values. *(FR-001, FR-009)*

### Permit (unchanged)

No schema or model changes. `affects_days` continues to be the single source of truth for "does this permit count toward the resident's limit" — set at creation time exactly as today (`exception_reason == "" && !resident.unlim_days`).

### Allowance Adjustment (new)

New table `quota_adjustment`.

| Field | Type | Notes |
|---|---|---|
| `id` | `SERIAL PRIMARY KEY` | |
| `resident_id` | `CHAR(8) REFERENCES resident(id) ON DELETE CASCADE NOT NULL` | the resident whose allowance this adjusts |
| `amount` | `SMALLINT NOT NULL` | signed; positive = grant, negative = reduction/offset. `NOT NULL`, and the service layer rejects `amount == 0` (a no-op adjustment isn't a meaningful audit entry) |
| `reason` | `TEXT NOT NULL` | required at the DB level; service layer additionally rejects empty/whitespace-only strings |
| `created_by_admin_id` | `TEXT REFERENCES admin(id)` | **nullable.** `NULL` only for the migration-generated reconciliation rows (see Decision 5 in research.md); every adjustment created through the API MUST have a non-null value, enforced in `app/quota_adjustment.go` |
| `created_at` | `TIMESTAMPTZ NOT NULL DEFAULT now()` | ordering field for chronological history (FR-007) |

**Validation rules** (enforced in `app/quota_adjustment.go`, mirroring how `app/permit.go` layers DB constraints with service-level checks):
- `resident_id` must reference an existing resident (`errs.NewNotFound("resident")` if not — matches existing repo error convention).
- `amount` must be a non-zero integer.
- `reason` must be non-empty after trimming whitespace.
- `created_by_admin_id` must be the authenticated admin's id (never client-supplied) for API-created adjustments; there is no code path in `app/` that allows a caller to set it to `NULL` — only the SQL migration inserts `NULL` rows, directly, once.

**State transitions**: none. Rows are immutable after insert — the repo interface exposes no `Update` or `Delete` method (structural enforcement of FR-004), and the API layer exposes no corresponding endpoints. A correction is always a new row.

**Relationships**:
- `quota_adjustment.resident_id` → `resident.id` (many adjustments per resident, cascade-deleted with the resident, matching existing `car`/`permit` FK behavior).
- `quota_adjustment.created_by_admin_id` → `admin.id` (nullable; many adjustments per admin).

## Derived value formulas (single source of truth, used by both validation and display per FR-008)

```
days_used(resident)          = SUM(permit.days) WHERE permit.resident_id = resident.id AND permit.affects_days = true
adjustment_total(resident)   = SUM(quota_adjustment.amount) WHERE quota_adjustment.resident_id = resident.id
effective_allowance(resident)= 20 + adjustment_total(resident)
```

`permit.days` here means `(end_ts - start_ts) / 86400`, i.e. the same day-count computation `util.GetAmtDays` already performs in Go for a single permit — the SQL aggregate is the set-based equivalent, not a new definition.

## Migration sequence

1. **`000007_quota_adjustment.up.sql`**:
   - `CREATE TABLE quota_adjustment (...)` as specified above.
   - Backfill: for every resident where `resident.amt_parking_days_used <> derived days_used`, `INSERT INTO quota_adjustment (resident_id, amount, reason, created_by_admin_id) VALUES (resident.id, resident.amt_parking_days_used - derived_days_used, 'Migration reconciliation: preserves previously-recorded allowance from legacy amt_parking_days_used field', NULL)`. Skipped for residents where the values already match (no-op adjustment avoided, per the `amount != 0` rule above).
   - `000007_quota_adjustment.down.sql`: drop the table (the backfilled rows are lost on rollback, same as any other migration rollback in this repo).

2. **`000008_drop_amt_parking_days_used.up.sql`**: `ALTER TABLE resident DROP COLUMN amt_parking_days_used`.
   - `000008_drop_amt_parking_days_used.down.sql`: re-add the column (`ADD COLUMN amt_parking_days_used SMALLINT NOT NULL DEFAULT 0`); note a rollback after this point cannot recover the original per-resident values (they now live only in `quota_adjustment`), which is acceptable since this mirrors how any destructive-column-drop migration behaves in this repo already.

3. `storage/psql/database.go`'s `CreateSchemas()` is updated to additionally apply `000007_quota_adjustment.up.sql` (and later `000008_drop_amt_parking_days_used.up.sql`) directly, without changing its existing `migrator.Migrate(1)` call — see research.md Decision 6 for why a simple version bump would incorrectly pull in the seed-data migrations too.
