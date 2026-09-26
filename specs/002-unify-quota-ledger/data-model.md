# Phase 1 Data Model: Unify Permit and Admin Quota Changes into One Ledger

## Entities

### `resident` (modified)

Drops the writable usage column entirely.

| Column | Change |
|---|---|
| `amt_parking_days_used` | **Removed.** Usage is never stored; always derived from `quota_ledger` (see `QuotaLedgerRepo.SelectBalances`, research.md §7). |

`unlim_days` is unchanged (still governs the "unlimited days" exemption, FR-007).

`models.Resident.AmtParkingDaysUsed *int` is removed from the struct; `ResidentService`/handlers instead attach a derived `DaysUsed int` / `EffectiveAllowance int` pair, presented as two distinct fields (spec FR-015) when returning a resident.

### `permit` (modified)

| Column | Change |
|---|---|
| `cancelled_ts` | **New**, `BIGINT NULL` (epoch seconds; matches `start_ts`/`end_ts`/`request_ts`). `NULL` = active/not cancelled. Set once, never cleared. |

Validation/lifecycle rules:
- A permit is created via `INSERT` exactly as today, plus a same-transaction `quota_ledger` insert (see below).
- "Deleting" a permit becomes: set `cancelled_ts = now()`, insert an offsetting `quota_ledger` row. The `permit` row is never physically deleted by this path (FR-016). `PermitRepo.Delete` is removed from the interface; a new `PermitRepo.Cancel(id int) error` (or the operation lives entirely inside `QuotaLedgerRepo`'s transactional method — see research.md §6) replaces it.
- `PermitRepo.Update` (license plate/color/make/model edits) is unaffected — it never touched day-count-relevant fields and still doesn't.
- Every "current state" read (`SelectWhere`, the resident/car active-permit-overlap checks in `app/permit.go`, all `GET /permits/*` list endpoints) implicitly filters `cancelled_ts IS NULL`. `GetOne(id)` does **not** filter it — a cancelled permit must remain individually fetchable so a `quota_ledger.permit_id` reference always resolves (FR-016).
- `PermitRepo.Reset()` (test-only full wipe) is unchanged.

### `quota_ledger` (new, replaces `quota_adjustment`)

```sql
CREATE TABLE quota_ledger (
    id                  BIGSERIAL PRIMARY KEY,
    resident_id         CHAR(8) NOT NULL REFERENCES resident(id) ON DELETE CASCADE,
    quota_year          INT NOT NULL,
    amount              SMALLINT NOT NULL,
    entry_type          TEXT NOT NULL CHECK (entry_type IN ('PERMIT', 'ADMIN_ADJUSTMENT')),
    permit_id           INT NULL REFERENCES permit(id),
    created_by_admin_id TEXT NULL REFERENCES admin(id),
    note                TEXT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX quota_ledger_resident_year_idx ON quota_ledger (resident_id, quota_year);
```

Field notes:

| Field | Meaning |
|---|---|
| `resident_id` | Whose usage this entry affects. |
| `quota_year` | The allowance period this entry counts against — computed at write time (research.md §3), **never** recomputed from `created_at` later (FR-011). For a `PERMIT` entry, this is the period in effect when the permit was created. For its cancellation's offsetting entry, this is copied from the original entry's `quota_year`, not the period current at cancellation time (FR-017, clarification Q2). |
| `amount` | Signed day count. Positive = adds to usage (a new permit, or a restrictive admin adjustment). Negative = removes from usage (a cancellation's offsetting entry, or a grant-style admin adjustment). |
| `entry_type` | `'PERMIT'` (system-generated, alongside a permit create/cancel) or `'ADMIN_ADJUSTMENT'` (admin-recorded, requires `note` and `created_by_admin_id`). |
| `permit_id` | Set only for `PERMIT` entries; `NULL` for `ADMIN_ADJUSTMENT`. References the permit whose creation or cancellation produced this row — always resolvable via `PermitRepo.GetOne`, even after the permit is cancelled. |
| `created_by_admin_id` | Set only for `ADMIN_ADJUSTMENT` entries. |
| `note` | Required (non-empty) for `ADMIN_ADJUSTMENT` entries (FR-005/edge case); `NULL` for `PERMIT` entries (the `permit_id` link is the explanation). |
| `created_at` | Wall-clock time the row was written. Informational/audit only — `quota_year`, not `created_at`, is authoritative for which period an entry counts against. |

Invariants (enforced in `QuotaLedgerRepo`, not by any DB-role privilege change — research.md §2):
- No `Update` or `Delete` method exists on `QuotaLedgerRepo`. Only `Create` (single entry, always inside the caller's transaction) and read methods.
- A resident's derived usage for a period = `SUM(amount) WHERE resident_id = ? AND quota_year = ?` (`COALESCE(..., 0)` when no rows exist — this is what makes the annual reset automatic, FR-012).
- A resident's effective allowance for a period = `20 + SUM(amount) WHERE resident_id = ? AND quota_year = ? AND entry_type = 'ADMIN_ADJUSTMENT'` (FR-006). Equivalently, "usage" already nets admin adjustments in; the two figures (FR-015) are reported separately to callers as described above.

### `admin` (unchanged)

Referenced by `quota_ledger.created_by_admin_id`; no schema change.

### `car` (unchanged — out of scope)

`car.amt_parking_days_used` and its increment/decrement calls in `app/permit.go` are left exactly as they are today; this feature does not touch the per-car counter (spec's Out of Scope).

## Removed

- `quota_adjustment` table (empty in every environment; dropped, not migrated — research.md §1).
- `models.QuotaAdjustment` struct and its constructor (`models/quota_adjustment.go`).
- `resident.amt_parking_days_used` column, `models.Resident.AmtParkingDaysUsed` field, `ResidentRepo.AddToAmtParkingDaysUsed`, and the `validateEditAmtDays` wiring inside `validator.EditResident` (the shared `validateEditAmtDays` function itself stays — `validator.EditCar` still uses it for the untouched per-car counter).
- `PermitRepo.Delete` (replaced by cancellation, see above).

## State Transitions

**Permit**: `(none)` → created (ledger `PERMIT` entry, `amount = +days`) → optionally cancelled (ledger `PERMIT` entry, `amount = -days`, same `quota_year` as creation). No other transitions; license-plate/color/make/model edits don't affect this state machine.

**Quota ledger entry**: `(none)` → created. That's the only transition — entries are immutable for their entire lifetime (FR-002).

## Validation Rules Summary (from spec Functional Requirements)

- `ADMIN_ADJUSTMENT` entries require a non-empty `note` (FR-005, edge case "adjustment without a reason").
- A new permit is accepted only if `derived_usage(resident, current_quota_year) + permit_days <= 20`, unless `resident.unlim_days` or the permit is exempt (`exception_reason != ""`) (FR-007) — checked inside the locked transaction (research.md §6) so concurrent requests can't both pass (FR-008).
- Cancelling a permit that was never counted (exempt, or belonging to an unlimited resident) does not produce a ledger entry — mirrors the "never counted, don't reverse what was never added" edge case.
