# Phase 0 Research: Unify Permit and Admin Quota Changes into One Ledger

## 1. Replace `quota_adjustment` with `quota_ledger`, and remove the old table

**Decision**: Drop the `quota_adjustment` table entirely (migration) and introduce a new `quota_ledger` table that records both permit-driven and admin-driven quota events, per spec FR-001/FR-014 and clarification session 2026-09-26.

**Rationale**: `quota_adjustment` is on `main` (merged from the superseded `001-quota-usage-separation` effort) but is empty in every environment — it was never wired into the app/api layers (tasks 4–5 that would have done so were never merged). There is no data to migrate out of it, so replacing it outright is strictly simpler than trying to evolve its schema in place, and avoids leaving a second, now-pointless table around (the exact "two sources of truth" ambiguity this feature exists to remove).

**Alternatives considered**: Renaming/altering `quota_adjustment` in place to add permit-linkage columns — rejected because the table's whole shape (adjustments only, no `entry_type`/`permit_id`) doesn't match the unified design, and there's no data to preserve by keeping the same physical table.

## 2. Enforcing ledger immutability (append-only)

**Decision**: Enforce "no UPDATE/DELETE, ever" at the application layer: `QuotaLedgerRepo` (new) exposes only `Create` and read methods — no `Update`/`Delete` method exists on the interface at all, so calling code cannot compile a mutation even by mistake. This mirrors the existing `PermitRepo`/`ResidentRepo`/`CarRepo` pattern of a narrow, purpose-built interface per entity.

**Rationale**: The vault-note design (`unifying-parkspots-day-usage-into-one-quota-ledger.md`) calls for revoking UPDATE/DELETE at the database-role level so the guarantee is "enforced, not just conventional." This repo's entire deployment (`config/postgres.go`) uses a single `DATABASE_URL` / single Postgres role for both migrations and the running app — there is no existing second, lower-privileged role to revoke from, and introducing one is an infrastructure change (provisioning, Heroku plan support, connection-string wiring) outside this codebase's current pattern and outside what this feature needs to touch. Application-layer enforcement (no mutation method exists) is the codebase-consistent equivalent given that constraint.

**Alternatives considered**: Provisioning a second, restricted Postgres role and `REVOKE UPDATE, DELETE ON quota_ledger FROM app_role` — noted as a valid future hardening step (left as a one-line callout in `data-model.md`), but not implemented now since it requires infrastructure this feature doesn't otherwise need and isn't reachable from a single-role `DATABASE_URL` setup without an unrelated ops change.

## 3. `quota_year` computation (no existing timezone concept)

**Decision**: `quota_year` is computed as `time.Now().UTC().Year()` at the moment a ledger entry is written (permit creation, permit cancellation, or admin adjustment), and stored explicitly on the row — never re-derived later from `created_at`.

**Rationale**: The codebase has no existing "organization canonical timezone" concept to defer to (grep across `config/`, `util/` found none); permit dates are already stored and compared as raw Unix timestamps (`start_ts`, `end_ts`, `request_ts` — see `storage/psql/permit_repo.go`), implicitly UTC/server-clock-based. UTC is the simplest choice consistent with that existing convention and needs no new configuration surface.

**Alternatives considered**: Adding a per-deployment timezone config — rejected as unjustified scope for this feature; nothing in the spec calls for anything other than a single, global allowance period boundary.

## 4. Permits are never hard-deleted (soft-cancel)

**Decision**: Add `permit.cancelled_ts BIGINT NULL` (epoch seconds, matching the table's existing `start_ts`/`end_ts`/`request_ts` convention). "Cancel" becomes `UPDATE permit SET cancelled_ts = $1 WHERE id = $2`, never `DELETE`. All existing read paths that represent "current state" (`SelectWhere` used for active-permit-overlap checks, the two-active-permits check, list/get-all endpoints) add an implicit `cancelled_ts IS NULL` filter so a cancelled permit behaves as if absent from every current-state view. `GetOne(id)` stays unfiltered, so a ledger entry's `permit_id` can always be resolved back to the permit record it refers to (per FR-016), including after cancellation.

**Rationale**: Directly implements the clarification answer (Question 1) and FR-016. Using the same epoch-int style as the permit table's other timestamp columns is more locally consistent than introducing a `TIMESTAMPTZ` column on this particular table (unlike `quota_ledger`, which follows the already-merged `quota_adjustment` table's `TIMESTAMPTZ` convention instead, since it's a fresh table with no legacy column style to match).

**Alternatives considered**: A boolean `is_cancelled` flag — rejected in favor of a nullable timestamp, since it captures "cancelled, and when" in one column with no loss of information, and this codebase already favors nullable timestamp/optional-pointer fields for similar optional facts (e.g., `exception_reason`).

## 5. Backfill granularity at cutover

**Decision**: The cutover migration inserts one `quota_ledger` row per existing qualifying permit (`entry_type='PERMIT'`, `permit_id` set, `quota_year` derived from that permit's own `request_ts`), then computes, per resident, the (small) residual difference between their old `resident.amt_parking_days_used` value and the sum just backfilled, and inserts at most one additional `entry_type='ADMIN_ADJUSTMENT'` row (`created_by_admin_id = NULL`, `note` explaining it's a migration reconciliation) to close that gap — mirroring the reconciliation SQL already present in `migrations/000007_quota_adjustment.up.sql`, adapted to per-permit backfill.

**Rationale**: Directly implements clarification Question 3/FR-013. Reuses the existing reconciliation-SQL shape from the (soon-removed) `000007` migration rather than inventing a new pattern.

## 6. Cross-entity transaction for permit create/cancel + ledger write

**Decision**: Introduce one new, narrowly-scoped SQL operation — not a general "every repo takes a `*sqlx.Tx`" refactor — that performs the whole permit-creation write path atomically:

```
BEGIN;
SELECT 1 FROM resident WHERE id = $1 FOR UPDATE;   -- lock this resident's row
SELECT COALESCE(SUM(amount), 0) FROM quota_ledger WHERE resident_id = $1 AND quota_year = $2;
-- app code checks sum + permitLength <= 20 (or unlimited/exempt)
INSERT INTO permit (...) VALUES (...) RETURNING id;
INSERT INTO quota_ledger (resident_id, quota_year, amount, entry_type, permit_id) VALUES ($1, $2, $3, 'PERMIT', <new permit id>);
COMMIT;
```

Permit cancellation follows the same lock-then-insert-then-commit shape (`UPDATE permit SET cancelled_ts = ...` + an offsetting `quota_ledger` insert, in one transaction, no need to re-check the limit since cancelling only ever decreases usage).

This lives on the new `QuotaLedgerRepo` (in `storage/psql`), which is handed the shared `*sqlx.DB` driver like every other repo, and opens/commits its own `*sqlx.Tx` internally for this one cross-table operation. `PermitService.Create`/`Delete` call it instead of the current separate `permitRepo.Create` + `residentRepo.AddToAmtParkingDaysUsed` calls.

**Rationale**: FR-008 (no double-booking under concurrent requests) cannot be met with the current codebase's separate, non-transactional repo calls (confirmed: `app/permit.go`'s `create()`/`Delete()` call `residentRepo.AddToAmtParkingDaysUsed` and `permitRepo.Create`/`Delete` as two independent, unsynchronized statements today — this is itself a latent race in the current code, not just a new requirement). `SELECT ... FOR UPDATE` on the resident row is the standard Postgres row-lock pattern for "read-then-conditionally-write" races and requires no new dependency (`lib/pq`/`sqlx` already support transactions via `driver.Beginx()`).

**Alternatives considered**: `SERIALIZABLE` isolation with retry-on-conflict (the vault note's other suggested option) — rejected as more complex to wire into this codebase's simple, non-retrying handler layer than an explicit row lock; `SELECT ... FOR UPDATE` gives the same correctness guarantee for this specific access pattern (single resident row, single writer path) without needing a retry loop. Refactoring every repo interface to accept a shared transaction executor — rejected as disproportionate to this feature; only the permit/ledger write path needs atomicity.

## 7. Effective allowance computation stays a single grouped query

**Decision**: `QuotaLedgerRepo` exposes `SelectBalances(residentIDs []string, quotaYear int) (map[string]int, error)` (or equivalent), built as one `SELECT resident_id, COALESCE(SUM(amount), 0) FROM quota_ledger WHERE quota_year = $1 GROUP BY resident_id` (optionally filtered to a set of IDs for a single page), used by `ResidentService.GetAll`/`GetOne` to attach each resident's derived usage — never a per-resident follow-up query. This directly satisfies FR-010/SC-003.

**Rationale**: Matches the vault note's core fix for the N+1 problem and requires no caching layer, consistent with the note's own conclusion that a live aggregate is cheap at this data scale (tens to low hundreds of ledger rows per resident per year).
