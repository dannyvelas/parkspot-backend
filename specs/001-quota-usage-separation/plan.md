# Implementation Plan: Separate Permit Usage from Admin-Editable Allowance

**Branch**: `001-quota-usage-separation` | **Date**: 2026-09-24 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-quota-usage-separation/spec.md`

## Summary

Residents' "days used" is currently a single writable integer column (`resident.amt_parking_days_used`) that the app increments/decrements on permit create/delete, but that admins can also overwrite directly via the resident edit endpoint. This makes it impossible to tell, after the fact, whether a resident's usage state is the result of a bug or a deliberate admin edit.

This plan removes that ambiguity by (1) deriving days-used at read/validation time as a live SUM over the resident's non-deleted, limit-counting permits — with no stored, writable copy of that number anywhere — and (2) introducing a new, append-only `quota_adjustment` ledger table that is the *only* way a resident's effective allowance can differ from the standard 20-day base. Permit-creation validation is rewritten to compare derived usage against `20 + SUM(quota_adjustment.amount)`. A one-time data migration converts every resident's pre-existing usage/limit mismatch into an explicit, auditable reconciliation adjustment, so no resident is newly blocked by this change.

## Technical Context

**Language/Version**: Go 1.24 (per `go.mod` / Heroku buildpack pin)

**Primary Dependencies**: `go-chi/chi` v5 (HTTP routing), `jmoiron/sqlx` + `Masterminds/squirrel` (query building over `database/sql`), `lib/pq` (Postgres driver), `golang-migrate/migrate` v4 (schema migrations), `golang-jwt/jwt` (session auth), `rs/zerolog` (logging)

**Storage**: PostgreSQL (existing tables: `admin`, `resident`, `car`, `permit`, `visitor`)

**Testing**: `stretchr/testify` (`suite` + `require`) with two established patterns in this repo: (a) real-Postgres integration tests via `testcontainers-go` + `storage/psql.NewSandboxDatabase()` (used by `app/permit_test.go`, `api/car_handler_test.go`), and (b) mock-repo unit tests via hand-written `storage/*_repo_mock.go` implementations of the repo interfaces (used by `app/resident_test.go`)

**Target Platform**: Linux server (Heroku-deployed HTTP API)

**Project Type**: Single backend service — layered architecture: `models/` (domain structs) → `storage/` (repo interfaces + mocks) → `storage/psql/` (Postgres implementations via squirrel/sqlx) → `app/` (business logic / validation services) → `api/` (chi HTTP handlers) → `main.go` (wiring)

**Performance Goals**: No stated targets; existing endpoints are simple CRUD over a small residential-community dataset (per-resident permit counts are small, typically well under a few hundred rows). A live `SUM(...)` aggregate scoped to one resident's permits is expected to be sub-millisecond and needs no caching.

**Constraints**: Must not change the per-car day-usage counter (`car.amt_parking_days_used`) — explicitly out of scope per spec clarification. Must not introduce a time-based reset of the 20-day limit. Existing residents' effective allowance must not shrink as a side effect of migration (FR-010).

**Scale/Scope**: Single residential-community backend; resident count and permits-per-resident are small (tens to low hundreds), consistent with existing `MaxLimit = 1000` pagination cap in `config/constants.go`.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

`.specify/memory/constitution.md` in this repository is still the unfilled template (no principles have been ratified — all fields are placeholder brackets). There are no project-specific constitutional gates to check against. This plan instead defaults to the general engineering defaults already stated in project guidance (YAGNI/simplicity, no speculative abstraction, match existing codebase conventions), which are reflected in the design decisions below (e.g., preferring a live aggregate query over a cached/recomputed column, reusing the existing repo/service/handler layering instead of introducing a new pattern).

**Result**: PASS (no gates defined; no violations to track in Complexity Tracking).

**Post-Design re-check** (after Phase 1 artifacts below): No new dependencies, services, or projects were introduced beyond one additional vertical slice following the codebase's existing pattern (see `research.md`, `data-model.md`). Still PASS.

## Project Structure

### Documentation (this feature)

```text
specs/001-quota-usage-separation/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/            # Phase 1 output (/speckit-plan command)
│   └── quota-adjustments.md
└── tasks.md               # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

This is a single Go backend service; the existing layout is reused as-is (no new top-level directories):

```text
migrations/
├── 000007_quota_adjustment.up.sql / .down.sql   # NEW: quota_adjustment table + reconciliation backfill
└── 000008_drop_amt_parking_days_used.up.sql / .down.sql  # NEW: drop the legacy writable column

models/
├── quota_adjustment.go        # NEW: QuotaAdjustment struct + constructor
├── resident.go                 # MODIFIED: remove AmtParkingDaysUsed field, add derived DaysUsed / EffectiveAllowance
└── permit.go                   # unchanged

storage/
├── quota_adjustment_repo.go        # NEW: repo interface (Create, SelectByResident, SelectSumByResident)
├── quota_adjustment_repo_mock.go   # NEW: mock impl, matching existing *_repo_mock.go convention
├── resident_repo.go                # MODIFIED: replace AddToAmtParkingDaysUsed with SelectDaysUsed
├── resident_repo_mock.go           # MODIFIED accordingly
└── psql/
    ├── quota_adjustment.go         # NEW: row struct + toModels()
    ├── quota_adjustment_repo.go    # NEW: squirrel/sqlx implementation
    ├── resident.go                  # MODIFIED: drop amt_parking_days_used column mapping
    ├── resident_repo.go             # MODIFIED: SelectDaysUsed via SUM query; remove AddToAmtParkingDaysUsed
    └── database.go                  # MODIFIED: expose QuotaAdjustmentRepo(); bump CreateSchemas() target version

app/
├── quota_adjustment.go         # NEW: QuotaAdjustmentService (validates reason/amount, records adjustment, lists history)
├── quota_adjustment_test.go    # NEW
├── permit.go                    # MODIFIED: validation uses derived days-used + effective allowance instead of stored column; remove Add/Subtract calls
├── permit_test.go               # MODIFIED: update fixtures/assertions for derived usage
├── resident.go                   # MODIFIED: remove AmtParkingDaysUsed from editable Update fields
├── resident_test.go              # MODIFIED accordingly
└── app.go                        # MODIFIED: wire QuotaAdjustmentService into App

models/validator/
└── resident.go                   # MODIFIED: drop validateEditAmtDays / AmtParkingDaysUsed validation from EditResident

api/
├── quota_adjustment_handler.go       # NEW: admin-only endpoints under a resident's adjustments
├── quota_adjustment_handler_test.go  # NEW
├── resident_handler.go                # MODIFIED: response no longer includes a writable amt_parking_days_used; exposes daysUsed/effectiveAllowance
└── server.go                           # MODIFIED: register new routes
```

**Structure Decision**: Extend the existing single-service layered structure (`models` → `storage` → `storage/psql` → `app` → `api`) with one new vertical slice (`quota_adjustment`) that mirrors how every other entity (e.g., `visitor`, `car`) is implemented in this codebase, plus targeted edits to the `resident` and `permit` slices to remove the writable usage counter and switch validation to derived values.

## Complexity Tracking

> No constitution gates were violated; this section is intentionally empty.
