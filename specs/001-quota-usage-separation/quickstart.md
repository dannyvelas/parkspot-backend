# Quickstart: Validating Quota/Usage Separation

Validates the feature end-to-end against a real local Postgres instance, using this repo's existing setup flow (see root `README.md`).

## Prerequisites

1. Docker running; `docker compose up -d` (starts local Postgres).
2. `make migrate_up` — applies all migrations, including the new `000007_quota_adjustment` and `000008_drop_amt_parking_days_used` from this feature.
3. `.env` present (`cp .env.example .env` if not already done).
4. `go run -v main.go` — starts the API on the configured port.
5. An admin session: `POST /api/login` with seeded admin credentials (see README: sample data password is `notapassword`), keeping the returned auth cookie/token for subsequent requests.
6. A seeded or newly created resident (`B` + 7 digits, e.g. `B1234567`) with no existing permits, for a clean starting state.

## Scenario 1 — Days-used is derived, not editable (User Story 1)

1. `GET /api/resident/B1234567` → confirm `daysUsed: 0`, `effectiveAllowance: 20`, and confirm the response has no `amtParkingDaysUsed` field at all.
2. `POST /api/permit` for that resident, 12 days (`startDate`/`endDate` 12 days apart), no exception reason.
3. `GET /api/resident/B1234567` again → `daysUsed: 12`.
4. `DELETE /api/permit/{id}` for the permit just created.
5. `GET /api/resident/B1234567` again → `daysUsed: 0`, with no manual step taken to reset it.
6. `PUT /api/resident` with body `{"id": "B1234567", "amtParkingDaysUsed": 999}` → expect this field to have no effect (resident's derived `daysUsed` is still driven solely by permits, not this request).

**Expected outcome**: `daysUsed` always exactly reflects step 2–4's permits; nothing in this scenario required a direct edit to succeed.

## Scenario 2 — Auditable allowance adjustment (User Story 2)

1. `POST /api/quota-adjustment` as admin: `{"residentID": "B1234567", "amount": 5, "reason": "Quickstart validation grant"}`.
2. `GET /api/resident/B1234567` → `effectiveAllowance: 25`, `daysUsed` unchanged from before this step.
3. `GET /api/resident/B1234567/quota-adjustments` → one entry with `amount: 5`, the reason above, `createdByAdminID` equal to the logged-in admin's id, and a `createdAt` timestamp.

**Expected outcome**: the grant is visible both as a changed allowance and as an explicit, attributed ledger entry.

## Scenario 3 — Permit validated against real usage + allowance (User Story 3)

Starting from a resident at 18 of 20 days used (create permits summing to 18 days first):

1. `POST /api/permit` requesting 3 more days → expect rejection (`400`, exceeds-limit error), since `18 + 3 = 21 > 20`.
2. `POST /api/quota-adjustment`: `{"residentID": "...", "amount": 5, "reason": "Quickstart validation grant"}`.
3. Retry the same 3-day `POST /api/permit` → expect success, since `18 + 3 = 21 <= 25`.

**Expected outcome**: the same request is deterministically rejected, then accepted, purely as a function of the recorded adjustment — no other state changed.

## Scenario 4 — Append-only ledger (clarified behavior)

1. Attempt any edit/delete HTTP call against the adjustment created in Scenario 2 (no such route exists — confirm `404 Not Found` from the router itself, not an app-level error).
2. `POST /api/quota-adjustment` with `{"residentID": "...", "amount": -5, "reason": "Correcting quickstart validation grant"}`.
3. `GET /api/resident/.../quota-adjustments` → both the original `+5` and the offsetting `-5` rows are present; `effectiveAllowance` is back to 20.

**Expected outcome**: history is never rewritten — a correction is always visible as a second entry.

## Scenario 5 — Migration reconciliation (FR-010)

Run against a database created *before* this feature's migrations (i.e., re-run `make migrate_down` back past `000006`, hand-set a resident's legacy `amt_parking_days_used` to a value inconsistent with their real permit total, then `make migrate_up`):

1. Before migrating: set resident `X`'s `amt_parking_days_used = 17` while they actually have permits summing to `20` days (the exact scenario from the original bug report).
2. Run `make migrate_up`.
3. `GET /api/resident/X/quota-adjustments` → a system-generated row exists with `amount: -3`, `createdByAdminID: null`, and a reason mentioning migration reconciliation.
4. `GET /api/resident/X` → `daysUsed: 20`, `effectiveAllowance: 17` — i.e., the resident's previously-effective limit (17) is preserved exactly, even though it's below the standard base of 20.

**Expected outcome**: no resident becomes newly blocked from anything they could do immediately before the migration ran.
