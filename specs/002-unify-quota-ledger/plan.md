# Implementation Plan: Unify Permit and Admin Quota Changes into One Ledger

**Branch**: `002-unify-quota-ledger` | **Date**: 2026-09-26 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/002-unify-quota-ledger/spec.md`

## Summary

Residents' "days used" is derived today from a single writable column (`resident.amt_parking_days_used`), incremented/decremented by permit create/delete but also directly overwritable by admins — the original auditability bug. The immediately-prior effort (`001-quota-usage-separation`, tasks 1-3 merged) started fixing this with a separate `quota_adjustment` table for admin overrides only, but that design still requires two independent queries per resident (permits + adjustments) to compute usage, and still treats permits as a source of truth for state accounting for/computed separately from admin corrections.

This plan replaces that in-progress design entirely with a single, unified, append-only `quota_ledger` table that is the *only* source of truth for every quota-affecting event — both permit-driven (creation/cancellation) and admin-driven (manual adjustment) — tagged with an explicit `quota_year`. A resident's usage and allowance become pure aggregates over this one table, computable for an entire page of residents in one grouped query (killing the N+1 problem), and the annual reset disappears entirely (a new year's balance is naturally zero, since no rows exist yet). Permits are never hard-deleted going forward (soft-cancel via `cancelled_ts`), so every ledger entry's originating permit stays permanently inspectable. The already-merged `quota_adjustment` table (empty, never wired into the app) is dropped rather than extended.

## Technical Context

**Language/Version**: Go 1.25 (per `go.mod`)

**Primary Dependencies**: `go-chi/chi` v5 (HTTP routing), `jmoiron/sqlx` + `Masterminds/squirrel` (query building over `database/sql`), `lib/pq` (Postgres driver), `golang-migrate/migrate` v4 (schema migrations), `golang-jwt/jwt` (session auth), `rs/zerolog` (logging)

**Storage**: PostgreSQL (existing tables: `admin`, `resident`, `car`, `permit`, `visitor`; this feature adds `quota_ledger`, removes `quota_adjustment`)

**Testing**: `stretchr/testify` (`suite` + `require`) with this repo's two established patterns: (a) real-Postgres integration tests via `testcontainers-go` + `storage/psql.NewSandboxDatabase()` (used by `app/permit_test.go`), and (b) mock-repo unit tests via hand-written `storage/*_repo_mock.go` implementations (used by `app/resident_test.go`). The new transactional lock-and-insert path (research.md §6) needs integration-test coverage (mocks can't meaningfully simulate row-level locking), so its concurrency test (Scenario 3.4 in `quickstart.md`) belongs in the `testcontainers-go` suite alongside `permit_test.go`.

**Target Platform**: Linux server (Heroku-deployed HTTP API)

**Project Type**: Single backend service — layered architecture: `models/` (domain structs) → `storage/` (repo interfaces + mocks) → `storage/psql/` (Postgres implementations via squirrel/sqlx) → `app/` (business logic / validation services) → `api/` (chi HTTP handlers) → `main.go` (wiring)

**Performance Goals**: FR-010/SC-003 — displaying a full page of residents' usage must issue a bounded number of queries regardless of page size (one grouped `SUM ... GROUP BY` query, not one per resident). No other stated throughput targets; data scale is a small residential community (tens to low hundreds of ledger rows per resident per year, per the 20-day cap).

**Constraints**: Must not change the per-car day-usage counter (`car.amt_parking_days_used`) — out of scope per spec. Ledger rows must never be updated or deleted once written (FR-002) — enforced at the application layer, since this deployment has only one Postgres role (research.md §2), not via a DB-level `REVOKE`. Permit creation/cancellation and their ledger entries must be atomic and safe under concurrent requests for the same resident (FR-008) — requires this codebase's first `*sqlx.Tx` + `SELECT ... FOR UPDATE` usage (research.md §6).

**Scale/Scope**: Single residential-community backend; resident count and permits-per-resident are small (tens to low hundreds), consistent with the existing `MaxLimit = 1000` pagination cap in `config/constants.go`.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

`.specify/memory/constitution.md` in this repository is still the unfilled template (no principles have been ratified — all fields are placeholder brackets). There are no project-specific constitutional gates to check against. This plan defaults to the general engineering defaults already stated in project guidance (YAGNI/simplicity, no speculative abstraction, match existing codebase conventions), reflected below in preferring a narrowly-scoped new transactional repo method over a repo-wide refactor, and application-layer immutability over standing up new database-role infrastructure this feature doesn't otherwise need.

**Result**: PASS (no gates defined; no violations to track in Complexity Tracking).

**Post-Design re-check** (after Phase 1 artifacts above): The one new pattern introduced — a repo method that opens its own `*sqlx.Tx` for a cross-table (`permit` + `quota_ledger`) atomic write — is scoped to exactly the one write path that needs it (research.md §6), not applied repo-wide. No new external dependencies, services, or projects. Still PASS.

## Project Structure

### Documentation (this feature)

```text
specs/002-unify-quota-ledger/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/            # Phase 1 output (/speckit-plan command)
│   └── quota-ledger.md
└── tasks.md               # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

Single Go backend service; existing layout reused as-is (no new top-level directories):

```text
migrations/
└── 000008_quota_ledger.up.sql / .down.sql
    # NEW: drops quota_adjustment (empty, superseded); creates quota_ledger + its
    # (resident_id, quota_year) index; adds permit.cancelled_ts; backfills one
    # PERMIT-type entry per existing qualifying permit plus, per resident, at
    # most one ADMIN_ADJUSTMENT-type reconciliation entry for any residual gap;
    # drops resident.amt_parking_days_used

models/
├── quota_ledger.go            # NEW: QuotaLedger struct + constructor (replaces models/quota_adjustment.go)
├── quota_adjustment.go         # REMOVED
├── resident.go                  # MODIFIED: remove AmtParkingDaysUsed field
└── permit.go                    # MODIFIED: add CancelledTS *int64 field; Equal() accounts for it

storage/
├── quota_ledger_repo.go            # NEW: repo interface (Create* transactional methods, SelectHistory, SelectBalances)
├── quota_ledger_repo_mock.go       # NEW: mock impl, matching existing *_repo_mock.go convention
├── permit_repo.go                   # MODIFIED: Delete → Cancel; SelectWhere/SelectCountWhere/GetOne semantics per data-model.md
├── resident_repo.go                 # MODIFIED: remove AddToAmtParkingDaysUsed
├── resident_repo_mock.go            # MODIFIED accordingly
└── psql/
    ├── quota_ledger.go              # NEW: row struct + toModels()
    ├── quota_ledger_repo.go         # NEW: squirrel/sqlx implementation, incl. the transactional
    │                                  create-permit-with-ledger-entry and cancel-permit-with-ledger-entry
    │                                  methods (research.md §6), and the grouped-balance query (research.md §7)
    ├── quota_adjustment.go          # REMOVED
    ├── quota_adjustment_repo.go     # REMOVED (never existed on main — superseded before being built)
    ├── permit_repo.go                # MODIFIED: Delete→Cancel; add cancelled_ts IS NULL filters; GetOne unfiltered
    ├── permit.go                     # MODIFIED: map cancelled_ts
    ├── resident.go                   # MODIFIED: drop amt_parking_days_used column mapping
    ├── resident_repo.go               # MODIFIED: remove AddToAmtParkingDaysUsed
    ├── database.go                    # MODIFIED: expose QuotaLedgerRepo(); bump CreateSchemas() target version
    └── test_helpers.go                 # MODIFIED: schemaOnlyMigrations references 000008 instead of 000007

app/
├── quota_ledger.go              # NEW: QuotaLedgerService (validates admin-adjustment reason/amount, records it,
│                                    lists a resident's unified history, computes daysUsed/effectiveAllowance
│                                    for one or many residents)
├── quota_ledger_test.go         # NEW
├── permit.go                     # MODIFIED: Create/Delete call QuotaLedgerService's transactional methods
│                                    instead of permitRepo.Create + residentRepo.AddToAmtParkingDaysUsed;
│                                    getAndValidateResident compares derived usage/allowance instead of the
│                                    stored column
├── permit_test.go                # MODIFIED: fixtures/assertions for derived usage; new concurrency test
│                                    (quickstart.md Scenario 3.4)
├── resident.go                    # MODIFIED: remove AmtParkingDaysUsed from editable Update fields;
│                                    GetAll/GetOne attach derived daysUsed/effectiveAllowance via
│                                    QuotaLedgerService
├── resident_test.go               # MODIFIED accordingly
└── app.go                          # MODIFIED: wire QuotaLedgerService into App

models/validator/
└── resident.go                    # MODIFIED: drop validateEditAmtDays wiring from EditResident
                                     # (validator.EditCar keeps using validateEditAmtDays — car counter is untouched)

api/
├── quota_ledger_handler.go           # NEW: admin-only POST (adjustment) + GET (history) endpoints
├── quota_ledger_handler_test.go       # NEW
├── permit_handler.go                   # MODIFIED: deleteOne() message unchanged; behavior now cancels
├── resident_handler.go                  # MODIFIED: response exposes daysUsed/effectiveAllowance, no
│                                          writable amtParkingDaysUsed
└── server.go                             # MODIFIED: register new routes (see contracts/quota-ledger.md)
```

**Structure Decision**: Extend the existing single-service layered structure (`models` → `storage` → `storage/psql` → `app` → `api`) with one new vertical slice (`quota_ledger`) mirroring how every other entity is implemented in this codebase, replacing the never-fully-built `quota_adjustment` slice from the superseded effort, plus targeted edits to the `permit` and `resident` slices to remove the writable usage counter, add soft-cancel, and switch validation/read paths to derived, ledger-backed values.

## Complexity Tracking

> No constitution gates were violated; this section is intentionally empty.
