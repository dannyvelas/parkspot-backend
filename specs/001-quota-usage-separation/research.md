# Phase 0 Research: Separate Permit Usage from Admin-Editable Allowance

All Technical Context fields were resolvable directly from the existing codebase (`go.mod`, `migrations/`, `storage/`, `app/`, `api/`) — no external unknowns remain. This document instead records the design decisions needed to satisfy the spec's functional requirements within this codebase's conventions.

## Decision 1: How is "days used" derived?

**Decision**: Compute days-used as a live SQL aggregate at read/validation time — `SELECT COALESCE(SUM((end_ts - start_ts) / 86400), 0) FROM permit WHERE resident_id = $1 AND affects_days = true`. No column stores this value; `resident.amt_parking_days_used` is dropped entirely (see Decision 4).

**Rationale**:
- The `permit.affects_days` boolean already encodes exactly "counts toward this resident's limit" (`app/permit.go`: `AffectsDays = ExceptionReason == "" && !resident.UnlimDays`), and permit deletes are hard deletes (`DELETE FROM permit WHERE id = $1`) — so "non-deleted, limit-counting permits" is simply `WHERE resident_id = ? AND affects_days = true` with no soft-delete or status filtering needed.
- Clarification confirmed days-used must be a lifetime running total including expired permits (no time-based reset), which a plain `SUM` over all matching rows already gives for free — no extra "is it still active" filtering logic is needed.
- Data volume is small (residential community; tens to low hundreds of permits per resident at most), so a scoped, indexed (`resident_id`) aggregate is effectively O(log n) + O(matching rows) and needs no caching.
- A stored/cached value — even one that's "only ever recomputed, never hand-edited" — still requires a write path that must stay perfectly synchronized with permit creates/deletes; any gap in that sync (a failed transaction, a direct DB fix, a future code path that forgets to recompute) reintroduces exactly the class of bug this feature exists to eliminate. A live, side-effect-free query has no synchronization state to drift.

**Alternatives considered**:
- *Cached column, recomputed on every permit create/delete*: rejected — reintroduces a writable column and a two-step consistency requirement (write permit, then write cache) that can partially fail; the current bug's root cause is exactly "days-used is a column someone can touch."
- *Database view (`CREATE VIEW resident_days_used AS ...`)*: functionally equivalent to a live query but adds a schema object with no behavior benefit in this codebase, where the query builder (squirrel) already composes ad hoc queries easily. Not adopted, but the underlying SQL is the same either way — a future change to a view is a non-breaking internal optimization if scale ever demands it.

## Decision 2: Where do allowance adjustments live?

**Decision**: New table `quota_adjustment` — `id SERIAL PRIMARY KEY`, `resident_id CHAR(8) REFERENCES resident(id) ON DELETE CASCADE NOT NULL`, `amount SMALLINT NOT NULL`, `reason TEXT NOT NULL`, `created_by_admin_id TEXT REFERENCES admin(id)` (nullable — see Decision 5), `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`. The repo exposes only `Create`, `SelectByResident` (chronological), and `SelectSumByResident` — no `Update`/`Delete`, matching FR-004's append-only requirement structurally rather than just by convention.

**Rationale**: Mirrors this codebase's existing per-entity table pattern (`visitor`, `car`) exactly — one table, one repo interface, one psql implementation, one service, one handler. `ON DELETE CASCADE` matches the existing `car`/`permit` FK behavior for resident deletion. Omitting `Update`/`Delete` from the repo interface (not just from the API) means the append-only guarantee can't be bypassed by a future handler that forgets to check a business rule — there is no code path capable of mutating a row after insert.

**Alternatives considered**:
- *Reuse `permit` table with a "virtual" adjustment row type*: rejected — conflates two different concepts (a parking permit vs. an allowance grant) in one table/model, which is exactly the kind of ambiguity this feature is trying to remove.
- *Soft-delete/versioned adjustments (`Update` allowed, old rows kept)*: rejected per clarification — corrections must be new offsetting rows, not edits to history.

## Decision 3: How is effective allowance computed and enforced?

**Decision**: `effective_allowance = config.MaxParkingDays (20) + SUM(quota_adjustment.amount WHERE resident_id = ?)`. `PermitService.Create`'s existing check (`app/permit.go: getAndValidateResident`) is rewritten to compare `days_used + permitLength` against this computed allowance instead of the stored `resident.AmtParkingDaysUsed`, preserving the existing bypass for `resident.UnlimDays` and exception permits.

**Rationale**: Directly implements spec FR-005/FR-006. Reuses the same "unlimited days / exception" bypass already present in `validateCar`/`getAndValidateResident`, so behavior for those two existing cases is unchanged — only the number being compared changes from a stored value to two derived ones.

## Decision 4: What happens to `resident.amt_parking_days_used`?

**Decision**: Drop the column entirely, in a separate migration (`000008_drop_amt_parking_days_used`) applied *after* the reconciliation backfill (`000007`) has safely captured its final value in `quota_adjustment` rows.

**Rationale**: A column that still exists but is merely "not supposed to be written to by convention" is exactly the fragile state that produced the original bug (nothing stopped an admin edit before either). Removing the column makes direct mutation structurally impossible, satisfying FR-001/FR-009 at the schema level, not just the application level. Splitting backfill and drop into two migrations keeps the destructive step (`DROP COLUMN`) isolated and easy to reason about/roll back independently of the additive step.

**Alternatives considered**: Keep the column read-only at the application layer only (remove it from `validator.EditResident` and the repo's `Update`, but leave the DB column and the `AddToAmtParkingDaysUsed` writes in place as a "cache"). Rejected for the same reason as Decision 1 — a value that is both stored and derived is a standing invitation for the two to drift, which is the entire problem statement.

## Decision 5: Who is attributed on the migration-generated reconciliation adjustments?

**Decision**: `created_by_admin_id` is nullable. The one-time reconciliation rows inserted by migration `000007` have `created_by_admin_id = NULL` and `reason = 'Migration reconciliation: preserves previously-recorded allowance from legacy amt_parking_days_used field'`. Every adjustment created through the application (FR-003) is still required to carry a non-null admin id — that requirement is enforced at the service layer (`app/quota_adjustment.go`), not the DB constraint, precisely so the one legitimate system-generated exception remains possible without weakening the column's semantics for admin-created rows.

**Rationale**: There is no human admin to attribute an automated migration step to, and inventing a synthetic "system" admin row would pollute the real `admin` table with a fictional account that could then be impersonated or confused with a real user. A nullable FK plus a clearly-labeled reason is simpler and matches how `permit.exception_reason` already uses nullability to distinguish a special case in this codebase.

**Alternatives considered**: Require a real admin to "claim" each reconciliation adjustment manually before this feature ships. Rejected — clarification (FR-010) requires this to be automatic and non-blocking; manual claiming would delay rollout for no auditability benefit (the migration's own execution is already the audit trail for these specific rows).

## Decision 6: Migration test harness compatibility

**Decision**: `storage/psql/database.go`'s `CreateSchemas()` currently hardcodes `migrator.Migrate(1)` — i.e., integration tests (`testcontainers`-backed `NewSandboxDatabase`) only ever apply migration `000001_schemas`, deliberately skipping the seed-data migrations `000002`–`000006`. `migrate.Migrate(N)` is not a "jump to version N" operation — it walks every intermediate version sequentially, so naively changing this to `migrator.Migrate(8)` would also apply seed migrations `000002`–`000006` along the way, inserting rows that would break every existing integration test's assumption of a clean, unseeded schema. Instead, `CreateSchemas()` keeps `migrator.Migrate(1)` unchanged, and separately applies each new *schema-only* migration's `.up.sql` file directly via `driver.Exec` (reading the file from the `migrations/` directory, the same one golang-migrate itself reads from), bypassing golang-migrate's version tracking entirely for these extra files. A small, explicit slice of schema-only migration filenames (starting with `000007_quota_adjustment.up.sql`, later also `000008_drop_amt_parking_days_used.up.sql`) is maintained in `storage/psql/database.go` for this purpose.

**Rationale**: Discovered by reading `storage/psql/database.go` and `storage/psql/test_helpers.go` — every repo-level and app-level integration test in this repo (`app/permit_test.go`, `api/car_handler_test.go`, etc.) depends on `CreateSchemas()` for its Postgres schema, so this is a required, non-optional part of the implementation, not an incidental detail. Executing the new files' SQL directly (rather than through golang-migrate) is safe specifically because they are schema-only DDL (plus a backfill `INSERT ... SELECT` that matches zero rows against an empty sandbox database) — there is no seed data to accidentally introduce. Production deploys (`make migrate_up`) are unaffected: they use the real, unmodified `golang-migrate` CLI against every file in order, so `000002`–`000008` are all applied normally there.

**Alternatives considered**:
- Insert the new table directly into `000001_schemas.up.sql`. Rejected — that migration has already been applied to any real deployed database; editing an already-applied migration file breaks `golang-migrate`'s checksum/ordering guarantees and is not how any other schema change in this repo's history has been done (every later table added its own numbered migration).
- Change `CreateSchemas()` to `migrator.Up()` (apply everything, including seeds) and update every affected test's assertions to tolerate the seeded rows. Rejected as a much larger, riskier change than this feature warrants — it would touch the setup/assertions of every existing integration test suite for a concern (test schema setup) that is orthogonal to this feature.
