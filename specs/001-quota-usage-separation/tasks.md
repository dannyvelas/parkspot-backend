# Tasks: Separate Permit Usage from Admin-Editable Allowance

**Input**: Design documents from `/specs/001-quota-usage-separation/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/quota-adjustments.md, quickstart.md

**Tests**: Not explicitly requested in the spec. Test tasks below are limited to (a) updating existing tests that would otherwise fail because production behavior changed under them, and (b) two small new integration tests that verify cross-story composition (US3, US4) where no other task already covers verification. No TDD/contract-test scaffolding was added beyond that.

## Ordering contract (read this first)

**This task list is a single, strictly ordered stack, not independent per-story tracks.** Every task is written so that, given all tasks before it are already merged to `main`, completing just that one task and merging it leaves the repository compiling, passing its existing test suite, and behaving correctly (no partial/broken state) — suitable as one standalone PR. Task IDs (`T001`, `T002`, ...) are also the required merge order. A `[P]` marker means the task's *files* don't overlap with the immediately-surrounding tasks so it could be *developed* concurrently with them, but it must still be **merged in numeric order** — merging `T005` before `T004` is not safe even though both are marked `[P]`, because `T005` depends on `T004` having already landed (see each task's "Depends on"). Where two `[P]` tasks have no dependency relationship to each other at all (e.g. `T006`/`T007`), they may be merged in either order relative to each other, but both still require everything below `T006` to already be in `main`.

User story labels (`[US1]`..`[US4]`) are for traceability back to `spec.md` only. Because this feature's four user stories share one small set of underlying mechanisms (a derived usage number and one allowance ledger), some stories are fully delivered by tasks filed under an earlier story's label — this is called out explicitly at each checkpoint below rather than duplicated as separate implementation tasks.

## Phase 1: Setup

No setup tasks are required. All dependencies this feature needs (`sqlx`, `squirrel`, `lib/pq`, `golang-migrate`, `testify`, `testcontainers-go`) are already present in `go.mod`, and no new tooling, linting, or project scaffolding is introduced.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Introduces the `quota_adjustment` ledger (table, model, repo) and the derived-days-used query capability that every user story depends on. Nothing in this phase changes any existing behavior — it is purely additive.

- [ ] T001 Add migration `migrations/000007_quota_adjustment.up.sql` / `migrations/000007_quota_adjustment.down.sql`. Up: `CREATE TABLE quota_adjustment (id SERIAL PRIMARY KEY, resident_id CHAR(8) REFERENCES resident(id) ON DELETE CASCADE NOT NULL, amount SMALLINT NOT NULL, reason TEXT NOT NULL, created_by_admin_id TEXT REFERENCES admin(id), created_at TIMESTAMPTZ NOT NULL DEFAULT now())`, wrapped in `BEGIN;`/`COMMIT;` matching `000001_schemas.up.sql`'s style, followed by the reconciliation backfill: `INSERT INTO quota_adjustment (resident_id, amount, reason) SELECT r.id, r.amt_parking_days_used - COALESCE(p.derived, 0), 'Migration reconciliation: preserves previously-recorded allowance from legacy amt_parking_days_used field' FROM resident r LEFT JOIN (SELECT resident_id, SUM((end_ts - start_ts) / 86400) AS derived FROM permit WHERE affects_days = true GROUP BY resident_id) p ON p.resident_id = r.id WHERE r.amt_parking_days_used <> COALESCE(p.derived, 0)`. Down: `DROP TABLE quota_adjustment`. Per data-model.md's migration sequence.

- [ ] T002 In `storage/psql/database.go`, update `CreateSchemas()` per research.md Decision 6: after the existing `migrator.Migrate(1)` call succeeds, read `../migrations/000007_quota_adjustment.up.sql` and execute its contents directly via `database.driver.Exec(string(sqlBytes))` (do not change the `Migrate(1)` call itself — it must stay as-is so seed migrations `000002`-`000006` continue to be skipped in tests). Introduce a small named slice/constant listing this filename so later tasks can append to it. Depends on: T001.

- [ ] T003 [P] Add `models/quota_adjustment.go`: `type QuotaAdjustment struct { ID int; ResidentID string; Amount int; Reason string; CreatedByAdminID *string; CreatedAt time.Time }` with matching `json` tags (`id`, `residentID`, `amount`, `reason`, `createdByAdminID`, `createdAt`), and a `NewQuotaAdjustment(residentID string, amount int, reason string, createdByAdminID *string) QuotaAdjustment` constructor (no ID/CreatedAt params — those are DB-generated), following the style of `models/permit.go`. `CreatedByAdminID` is a pointer specifically so it can be `nil` for the migration-generated rows from T001. Depends on: none.

- [ ] T004 In `storage/quota_adjustment_repo.go` (new file), define `type QuotaAdjustmentRepo interface { Create(adjustment models.QuotaAdjustment) (models.QuotaAdjustment, error); SelectByResident(residentID string) ([]models.QuotaAdjustment, error); SelectSumByResident(residentID string) (int, error); Reset() error }`, matching the shape of `storage/resident_repo.go`. Deliberately no `Update`/`Delete` method (structural enforcement of the append-only rule from FR-004/data-model.md). Depends on: T003.

- [ ] T005 [P] Add `storage/psql/quota_adjustment.go`: unexported row struct mirroring `quota_adjustment`'s columns (`db` tags matching `storage/psql/permit.go`'s style) plus a `toModels()`/slice-`toModels()` conversion to `models.QuotaAdjustment`, converting the nullable `created_by_admin_id` column to `*string` via `sql.NullString`. Depends on: T003.

- [ ] T006 Add `storage/psql/quota_adjustment_repo.go`: `NewQuotaAdjustmentRepo(driver *sqlx.DB) storage.QuotaAdjustmentRepo` plus implementations of `Create` (INSERT with `Suffix("RETURNING *")`, scanned into the row struct), `SelectByResident` (`SELECT * FROM quota_adjustment WHERE resident_id = $1 ORDER BY created_at ASC`), `SelectSumByResident` (`SELECT COALESCE(SUM(amount), 0) FROM quota_adjustment WHERE resident_id = $1`), and `Reset` (`DELETE FROM quota_adjustment`), following `storage/psql/permit_repo.go`'s error-wrapping conventions (`errs.ErrDBBuildingQuery`, `errs.ErrDBExec`, `errs.ErrDBQuery`). Depends on: T001, T004, T005.

- [ ] T007 [P] Add `storage/quota_adjustment_repo_mock.go`: `QuotaAdjustmentRepoMock` implementing `storage.QuotaAdjustmentRepo` in-memory (slice-backed), following `storage/resident_repo_mock.go`'s pattern — `Create` appends and assigns an incrementing id + `time.Now()`, `SelectByResident` filters and preserves insertion order, `SelectSumByResident` sums matching rows, `Reset` clears the slice. Depends on: T004.

- [ ] T008 Add `QuotaAdjustmentRepo() storage.QuotaAdjustmentRepo` to the `storage.Database` interface in `storage/database.go`, and implement it in `storage/psql/database.go` (new `quotaAdjustmentRepo` field on the `Database` struct, wired in `NewDatabase` via `NewQuotaAdjustmentRepo(driver)`, exposed via a `QuotaAdjustmentRepo()` accessor method) — both files in this one task, since Go requires the interface and its implementer to change together to keep the build green. Depends on: T006.

- [ ] T009 [P] Add `SelectSumDaysByResident(residentID string) (int, error)` to the `PermitRepo` interface in `storage/permit_repo.go`, and implement it in `storage/psql/permit_repo.go` as `SELECT COALESCE(SUM((end_ts - start_ts) / 86400), 0) FROM permit WHERE resident_id = $1 AND affects_days = true` (both files in this one task, same interface/implementer reasoning as T008; there is no `PermitRepo` mock in this codebase, so no third file to update). Depends on: none (independent of the T001-T008 chain — different table, different files).

**Checkpoint**: `quota_adjustment` exists and is fully queryable/writable through the repo layer; derived days-used is queryable through `PermitRepo`. Nothing in the application yet reads or writes any of this — no behavior has changed.

---

## Phase 3: User Story 1 - Trustworthy, Always-Accurate Usage Tracking (Priority: P1) 🎯 MVP

**Goal**: A resident's days-used is always exactly `SUM` of their limit-counting permits, is shown alongside their real effective allowance, and can no longer be set directly by anyone. This phase also implements the actual permit-creation gating logic (US3's mechanism — see the note at the end of this phase).

**Independent Test**: Follow quickstart.md Scenarios 1 and 3 — create/delete permits and confirm `daysUsed` tracks them exactly with no manual step; confirm a permit request is rejected/accepted based on real usage vs. allowance.

- [ ] T010 [US1] In `models/resident.go`, add two new fields to `Resident` alongside the existing `AmtParkingDaysUsed *int` (do not remove it yet): `DaysUsed int` (`json:"daysUsed"`) and `EffectiveAllowance int` (`json:"effectiveAllowance"`). These are populated by the service layer in T011/T012, never by a repo scan — do not add `db` tags or wire them into `storage/psql/resident.go`'s row struct. Depends on: none.

- [ ] T011 [US1] In `app/resident.go`: add `permitRepo storage.PermitRepo` and `quotaAdjustmentRepo storage.QuotaAdjustmentRepo` fields to `ResidentService`, update `NewResidentService` to accept them as new parameters, and in `GetOne` and `GetAll`, after loading each resident, populate `resident.DaysUsed = permitRepo.SelectSumDaysByResident(resident.ID)` and `resident.EffectiveAllowance = config.MaxParkingDays + quotaAdjustmentRepo.SelectSumByResident(resident.ID)` (per data-model.md's formulas). `GetAll` calls both per resident in its result page (no batching — acceptable at this project's scale per plan.md's Performance Goals; do not add a batched/joined query). Update the one call site in `app/app.go`'s `NewApp` to pass `database.PermitRepo()` and `database.QuotaAdjustmentRepo()`. Depends on: T008, T009, T010.

- [ ] T012 [US1] In `app/permit.go`: add a `quotaAdjustmentRepo storage.QuotaAdjustmentRepo` field to `PermitService` and a new parameter to `NewPermitService`; update the one call site in `app/app.go`'s `NewApp` accordingly. Rewrite `getAndValidateResident` (around line 247-256) to replace the `*resident.AmtParkingDaysUsed` checks with: `daysUsed := s.permitRepo.SelectSumDaysByResident(resident.ID)`, `effectiveAllowance := config.MaxParkingDays + s.quotaAdjustmentRepo.SelectSumByResident(resident.ID)`, then `if daysUsed >= effectiveAllowance { return errs.EntityDaysTooLong("resident", daysUsed) } else if daysUsed+permitLength > effectiveAllowance { return errs.PermitPlusEntityDaysTooLong("resident", daysUsed) }` — same two existing error constructors, now fed derived values. Remove the two `s.residentRepo.AddToAmtParkingDaysUsed(...)` calls in `Delete` (line ~79) and `Create` (line ~288) entirely — usage is derived, so nothing needs incrementing/decrementing on permit create/delete (leave the paired `s.carService.carRepo.AddToAmtParkingDaysUsed(...)` calls untouched; the per-car counter is explicitly out of scope). Update `app/permit_test.go`: the `shouldAddDays`/`amtDaysAddedToRes` assertions (around lines 250-292) that check `AmtParkingDaysUsed` deltas must instead assert on `suite.permitService.permitRepo.SelectSumDaysByResident(...)` (or an equivalent derived check) before/after create and delete. Depends on: T008, T009.

- [ ] T013 [P] [US1] Stop `AmtParkingDaysUsed` from being admin-editable (FR-009), in three files: (1) `models/validator/resident.go` — remove `validateAmtDaysFn`/`validateEditAmtDays` field and its call in `Run`, so `EditResident` no longer validates or expects this field; (2) `app/resident.go`'s `Update` — remove `desiredResident.AmtParkingDaysUsed == nil` from the "all edit fields empty" check (line ~65) and its mention in the `errs.AllEditFieldsEmpty(...)` field list; (3) `storage/psql/resident_repo.go`'s `Update` — remove the `if residentFields.AmtParkingDaysUsed != nil { ... Set("amt_parking_days_used", ...) }` block (lines ~171-173). Update `app/resident_test.go`'s edit-fields test table (line ~95) to remove the `"amtParkingDaysUsed"` case (it's no longer a settable field, so there's nothing meaningful left to assert there). Depends on: T010 (field must still exist on the struct at this point — it does, this task only stops it being *written*, T014 removes it).

- [ ] T014 [US1] Remove `AmtParkingDaysUsed` and `AddToAmtParkingDaysUsed` entirely now that nothing reads or writes them: `models/resident.go` (remove the field, remove the `amtParkingDaysUsed int` parameter from `NewResident` and its use), `storage/psql/resident.go` (remove the row-struct field, its position in the `residentSelect` column list in `storage/psql/resident_repo.go`'s `NewResidentRepo`, and the corresponding argument in `models.NewResident(...)`'s call in `toModels()`), `storage/psql/resident_repo.go`'s `Create` (remove `amt_parking_days_used` from the `SetMap`, which — since `Create` never set it explicitly before either — should already default to the column's own `DEFAULT 0`; verify `NewResident`'s removed parameter doesn't break `Create`'s call site), `storage/resident_repo.go` (remove `AddToAmtParkingDaysUsed` from the interface), `storage/psql/resident_repo.go` (remove its implementation), `storage/resident_repo_mock.go` (remove its implementation and the `AmtParkingDaysUsed` handling in mock `Update`/`Create`), `models/test_helpers.go` (remove `AmtParkingDaysUsed: util.ToPtr(0)` from both `TestResident` and `TestResidentUnlimDays` literals). Depends on: T012, T013.

- [ ] T015 [US1] Add migration `migrations/000008_drop_amt_parking_days_used.up.sql` (`ALTER TABLE resident DROP COLUMN amt_parking_days_used;`, wrapped in `BEGIN;`/`COMMIT;`) and `migrations/000008_drop_amt_parking_days_used.down.sql` (`ALTER TABLE resident ADD COLUMN amt_parking_days_used SMALLINT NOT NULL DEFAULT 0;`). Update `storage/psql/database.go`'s `CreateSchemas()` (from T002) to also apply `000008_drop_amt_parking_days_used.up.sql`'s contents the same way it applies `000007`'s. Depends on: T014.

**Checkpoint**: User Story 1 is fully delivered — `daysUsed`/`effectiveAllowance` are always correct and derived, the old column is gone at both the Go and schema level, and it is structurally impossible for anyone to hand-edit usage again. **User Story 3's core requirement (FR-006 — permit creation validated against real usage + allowance) is also already fully delivered by T012** — no additional production-code task exists for it later; see the note at Phase 5.

---

## Phase 4: User Story 2 - Explicit, Auditable Allowance Adjustments (Priority: P1)

**Goal**: Admins can record and view allowance adjustments over HTTP. This phase depends only on the Foundational phase (T001-T009) and is independent of Phase 3 — it can be built and merged before, after, or interleaved with Phase 3's tasks.

**Independent Test**: Follow quickstart.md Scenario 2 (and Scenario 4 for the append-only guarantee) — record an adjustment, confirm it appears in history with full attribution, and confirm no edit/delete route exists.

- [ ] T016 [US2] Add `app/quota_adjustment.go`: `QuotaAdjustmentService` wrapping a `storage.QuotaAdjustmentRepo` (and a `storage.ResidentRepo` to validate resident existence), with `NewQuotaAdjustmentService(...)`, `Create(adjustment models.QuotaAdjustment) (models.QuotaAdjustment, error)` (validates: resident exists via `residentRepo.SelectWhere` → `errs.NewNotFound("resident")` if not found; `amount != 0` → `errs.InvalidFields("amount")`; non-empty, non-whitespace `reason` → `errs.InvalidFields("reason")`; then delegates to the repo), and `GetHistory(residentID string) ([]models.QuotaAdjustment, error)` (validates `residentID != ""` → `errs.MissingIDField`, then delegates to `SelectByResident`). Follows `app/permit.go`'s service-layer validation style. Depends on: T003, T004.

- [ ] T017 [US2] In `app/app.go`, add a `QuotaAdjustmentService` field to `App` and wire it in `NewApp` via `NewQuotaAdjustmentService(database.QuotaAdjustmentRepo(), database.ResidentRepo())`. Depends on: T008, T016.

- [ ] T018 [P] [US2] Add `api/quota_adjustment_handler.go`: `quotaAdjustmentHandler` wrapping `app.QuotaAdjustmentService`, with `newQuotaAdjustmentHandler(...)`, a `create()` handler (decodes body into `models.QuotaAdjustment`, overwrites `CreatedByAdminID` with the authenticated admin's id from `ctxGetAccessPayload` — never trusting a client-supplied value, per contracts/quota-adjustments.md — then calls the service and responds `200` with the created record), and a `getHistory()` handler (reads `{id}` via `chi.URLParam`, calls `GetHistory`, responds `200` with the list). Follows `api/permit_handler.go`'s structure exactly. Depends on: T016.

- [ ] T019 [US2] In `api/server.go`, instantiate `quotaAdjustmentHandler := newQuotaAdjustmentHandler(app.QuotaAdjustmentService)` and register `adminRouter.Post("/quota-adjustment", quotaAdjustmentHandler.create())` and `adminRouter.Get("/resident/{id}/quota-adjustments", quotaAdjustmentHandler.getHistory())` inside the existing `adminRouter` group (`AdminRole`-only, matching `POST /resident`'s auth level). Depends on: T017, T018.

**Checkpoint**: User Story 2 is fully delivered — adjustments are createable and viewable over HTTP, admin-only, append-only (no edit/delete route exists anywhere in `api/server.go`). **User Story 4's core requirement (FR-007 — chronological adjustment history) is also already fully delivered by T019**, since `SelectByResident` (T006) already orders by `created_at ASC`; see the note at Phase 6.

---

## Phase 5: User Story 3 - Permit Creation Validated Against Real Usage and Allowance (Priority: P2)

**Goal / already delivered**: This story's production behavior (FR-006) was implemented in **T012** as part of Phase 3, because in this codebase permit-creation validation and days-used derivation are the same code path — there is no separate "validation" component to build. This phase adds one integration test that specifically exercises US3's own acceptance scenario end-to-end (reject → admin grants adjustment via the API → same request now succeeds), which requires both Phase 3 and Phase 4 to be merged first.

- [ ] T020 [US3] Add a test to `app/permit_test.go` (new suite method, e.g. `TestCreate_RejectedThenAcceptedAfterAdjustment`): create permits for a resident totaling 18 days, attempt an additional 3-day permit and assert it's rejected with `errs.PermitPlusEntityDaysTooLong`-shaped error, call `suite.app.QuotaAdjustmentService.Create(...)` (or the equivalent service directly, matching how this suite already constructs services) with a `+5` adjustment, retry the same 3-day permit request, and assert it now succeeds. Depends on: T012, T019.

**Checkpoint**: User Story 3 is verified end-to-end on top of Stories 1 and 2.

---

## Phase 6: User Story 4 - Historical Adjustment Visibility for Investigations (Priority: P3)

**Goal / already delivered**: This story's production behavior (FR-007) was implemented in **T006** (chronological `SelectByResident`) and exposed in **T019** (`GET /resident/{id}/quota-adjustments`) as part of Phase 4. This phase adds one integration test confirming multiple adjustments compose correctly into a readable history, which only needs Phase 4.

- [ ] T021 [US4] Add a test (`app/quota_adjustment_test.go`, new file, following `app/permit_test.go`'s `testcontainers`-backed suite pattern) that creates two adjustments for the same resident at different times with distinct reasons/admins, calls `GetHistory`, and asserts both are present, in chronological order, with correct `amount`/`reason`/`createdByAdminID`/`createdAt` on each. Also asserts a resident with zero adjustments returns an empty (not `nil`-erroring) slice and an `effectiveAllowance` equal to `config.MaxParkingDays`. Depends on: T019.

**Checkpoint**: All four user stories are independently verifiable and demonstrated working together.

---

## Phase 7: Polish & Cross-Cutting Concerns

- [ ] T022 [P] Grep the repository for any remaining `AmtParkingDaysUsed` / `amt_parking_days_used` references (`grep -rn "AmtParkingDaysUsed\|amt_parking_days_used" --include="*.go" .`) and confirm the only hits left are in `car.go`/`car_repo.go`/`models/validator/car.go` files (the per-car counter, explicitly out of scope per spec's "Out of Scope" section) — no leftover resident-side references. This is a verification-only task; if it finds something, that's a signal an earlier task (T012-T015) was incomplete, not new work to invent. Depends on: T015.

- [ ] T023 Run through `quickstart.md` Scenarios 1-5 against a local `docker compose up -d` + `make migrate_up` Postgres instance and the running service, confirming each scenario's expected outcome. Depends on: T020, T021, T022.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: none — empty.
- **Foundational (Phase 2, T001-T009)**: no dependencies on other phases; BLOCKS every later phase.
- **User Story 1 (Phase 3, T010-T015)**: depends on Phase 2 only (specifically T008, T009).
- **User Story 2 (Phase 4, T016-T019)**: depends on Phase 2 only (specifically T003, T004, T008) — **independent of Phase 3**. Both phases touch `app/app.go` (T011/T012 in Phase 3, T017 in Phase 4), so whichever is merged second will need a small rebase there; that is normal stacking, not a broken intermediate state.
- **User Story 3 (Phase 5, T020)**: depends on both Phase 3 (T012) and Phase 4 (T019), since it verifies their composition.
- **User Story 4 (Phase 6, T021)**: depends on Phase 4 (T019) only.
- **Polish (Phase 7, T022-T023)**: depends on everything above.

### Strict merge order

T001 → T002 → T003 → T004 → T005 → T006 → T007 → T008 → T009 → T010 → T011 → T012 → T013 → T014 → T015 → T016 → T017 → T018 → T019 → T020 → T021 → T022 → T023.

(T003/T005/T007/T009/T013/T018/T022 are marked `[P]` because their *files* don't overlap with their immediate neighbors and they could be developed concurrently with them — but per the Ordering Contract above, PRs still merge in this numeric sequence.)

### MVP scope

User Story 1 (through **T015**) is the smallest slice that fixes the bug described in the original request: usage becomes derived and unforgeable, and permit creation is validated correctly. User Stories 2-4 (adjustments ledger + its HTTP surface + verification) can ship as a fast-follow stack of PRs afterward — or, since Phase 4 only depends on Phase 2, its tasks (T016-T019) could equally be stacked *before* T010-T015 if an admin-facing "grant extra days" capability is needed sooner than the full usage-derivation cleanup.

---

## Notes

- Every task lists explicit file paths and, where non-obvious, the exact SQL/Go signatures to use — this is intentional so each task is self-contained enough to hand to an implementer (human or LLM) without re-deriving decisions already made in `research.md`/`data-model.md`.
- Commit and open a PR after each task; do not batch multiple task IDs into one PR — that is the entire point of this task breakdown per the user's stacking requirement.
- If `/speckit-implement` is used to execute this list, confirm it processes tasks in ascending numeric order and opens one PR per task, consistent with the Ordering Contract above.
