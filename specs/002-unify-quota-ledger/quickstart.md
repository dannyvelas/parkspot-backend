# Quickstart: Validating the Unified Quota Ledger

Validates the feature end-to-end against a real local Postgres instance, using this repo's existing setup flow (see root `README.md`).

## Prerequisites

1. Docker running; `docker compose up -d` (starts local Postgres).
2. `make migrate_up` — applies all migrations, including this feature's new `000008_quota_ledger` (drops `quota_adjustment`, adds `quota_ledger` and `permit.cancelled_ts`, backfills history, drops `resident.amt_parking_days_used`).
3. `.env` present (`cp .env.example .env` if not already done).
4. `go run -v main.go` — starts the API on the configured port.
5. An admin session: `POST /api/login` with seeded admin credentials (see README: sample data password is `notapassword`).
6. A seeded or newly created resident (`B` + 7 digits, e.g. `B1234567`) with no existing permits, for a clean starting state.

## Scenario 1 — One unified history for every quota change (User Story 1)

1. `GET /api/resident/B1234567` → confirm `daysUsed: 0`, `effectiveAllowance: 20`, and no `amtParkingDaysUsed` field at all.
2. `POST /api/permit` for that resident, 3 days, no exception reason. Note the returned permit `id`.
3. `GET /api/resident/B1234567/quota-history` → one `PERMIT` entry, `amount: 3`, `permitID` matching step 2.
4. `DELETE /api/permit/{id}` for that permit.
5. `GET /api/resident/B1234567` → `daysUsed: 0` again, with no manual step taken.
6. `GET /api/resident/B1234567/quota-history` → **two** entries now: the original `+3` (unchanged) and a new offsetting `-3`, both referencing the same `permitID`.
7. `GET /api/permit/{id}` (the cancelled permit) → still returns `200` with the permit's data (not `404`), confirming it was soft-cancelled, not deleted.
8. `POST /api/quota-adjustment` as admin: `{"residentID": "B1234567", "amount": 5, "reason": "Quickstart validation grant"}`.
9. `GET /api/resident/B1234567/quota-history` → the two `PERMIT` entries from before, plus the new `ADMIN_ADJUSTMENT` entry, all in one chronological list.
10. Attempt any edit/delete HTTP call against any of these entries → no such route exists (`404` from the router itself).

**Expected outcome**: every event — permit creation, permit cancellation, admin adjustment — appears in one combined, append-only history, each attributed to its real cause.

## Scenario 2 — Fast usage totals for a resident list (User Story 2)

1. Create permits and/or admin adjustments for several residents.
2. `GET /api/residents` (paginated list) → confirm every resident's `daysUsed`/`effectiveAllowance` is correct and matches what Scenario 1's per-resident checks would show individually.
3. Confirm (e.g., via query logging, or reading `ResidentService.GetAll`) that computing the whole page issues one grouped query against `quota_ledger`, not one query per resident (FR-010).

**Expected outcome**: list totals are correct and the number of queries does not grow with the number of residents shown.

## Scenario 3 — Permit requests judged against the same total (User Story 3)

Starting from a resident at 18 of 20 days used (create permits summing to 18 days first):

1. `POST /api/permit` requesting 3 more days → expect rejection (`400`, exceeds-limit error), since `18 + 3 = 21 > 20`.
2. `POST /api/quota-adjustment`: `{"residentID": "...", "amount": 5, "reason": "Quickstart validation grant"}`.
3. Retry the same 3-day `POST /api/permit` → expect success, since `18 + 3 = 21 <= 25`.
4. Fire two concurrent `POST /api/permit` requests for the same resident, each large enough that only one could fit within their remaining allowance → confirm exactly one succeeds and one is rejected (FR-008; validates the row-lock in research.md §6).

**Expected outcome**: acceptance/rejection is a deterministic function of the ledger total, and concurrent requests can't both succeed when only one should.

## Scenario 4 — Automatic renewal at a period boundary (User Story 4)

1. Take a resident with a fully-used allowance in the current `quotaYear`.
2. Without running any script, job, or manual step, query their derived usage for `quotaYear + 1` (e.g., by inserting a test permit dated so its creation falls after simulating the year boundary, or by directly querying `quota_ledger` for a `quota_year` that has no rows yet) → confirm it reads `0`.
3. Confirm the prior year's entries are untouched and still visible via `quota-history`.

**Expected outcome**: a new period starts at zero automatically — no reset job exists to run or forget.

## Scenario 5 — Cutover backfill (FR-013)

Run against a database created *before* this feature's migration (i.e., re-run `make migrate_down` back past `000007`, seed permits and a stale `amt_parking_days_used`, then `make migrate_up`):

1. Before migrating: resident `X` has permits summing to `20` days, but `amt_parking_days_used = 17` (the exact scenario from the original bug report).
2. Run `make migrate_up`.
3. `GET /api/resident/X/quota-history` → one `PERMIT` entry per existing qualifying permit (each with its own `permitID` and the `quotaYear` derived from that permit's original `request_ts`), plus exactly one `ADMIN_ADJUSTMENT` reconciliation entry (`createdByAdminID: null`, reason mentioning migration reconciliation) closing the residual gap.
4. `GET /api/resident/X` → `daysUsed: 20`, `effectiveAllowance: 17` — the resident's previously-effective limit is preserved exactly, even though it's below the standard base of 20.

**Expected outcome**: pre-cutover activity is just as individually auditable as anything created afterward, and no resident becomes newly blocked from anything they could do immediately before the migration ran.
