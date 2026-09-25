# Contract: Allowance Adjustments

Follows this codebase's existing route conventions exactly: flat, singular-noun paths under `/api`, JSON request/response bodies, errors returned via the existing `errs.APIErr` → `respondError` convention (HTTP status + message), auth via the existing `middleware.authenticate(role...)` chi middleware.

## `POST /api/quota-adjustment`

**Auth**: `AdminRole` only (same group as `POST /api/resident`).

**Request body**:

```json
{
  "residentID": "B1234567",
  "amount": 5,
  "reason": "Board-approved extension for holiday guests, case #482"
}
```

**Behavior**:
1. Reject if `residentID` does not reference an existing resident → `404` (`errs.NewNotFound("resident")`, matching existing convention).
2. Reject if `amount == 0` → `400` (new `errs.InvalidFields`-style error: "amount must be non-zero").
3. Reject if `reason` is empty/whitespace-only → `400` (new error: "reason must not be empty").
4. `createdByAdminID` is always taken from the authenticated admin's access-token payload (`ctxGetAccessPayload`), never from the request body — mirrors how `POST /api/permit` derives `residentID` from the token for resident callers.
5. On success, insert the row and return it.

**Response** (`200`):

```json
{
  "id": 42,
  "residentID": "B1234567",
  "amount": 5,
  "reason": "Board-approved extension for holiday guests, case #482",
  "createdByAdminID": "A0000001",
  "createdAt": "2026-09-24T18:04:00Z"
}
```

**Not supported**: `PUT`/`PATCH`/`DELETE` on this resource. There is intentionally no endpoint to edit or remove an adjustment (FR-004) — corrections are made by `POST`-ing a new, offsetting adjustment.

## `GET /api/resident/{id}/quota-adjustments`

**Auth**: `AdminRole` only.

**Behavior**: Returns the resident's full adjustment history, ordered chronologically (oldest first, matching the "in chronological order" wording of FR-007 / US4).

**Response** (`200`):

```json
[
  {
    "id": 17,
    "residentID": "B1234567",
    "amount": -3,
    "reason": "Migration reconciliation: preserves previously-recorded allowance from legacy amt_parking_days_used field",
    "createdByAdminID": null,
    "createdAt": "2026-09-24T00:00:00Z"
  },
  {
    "id": 42,
    "residentID": "B1234567",
    "amount": 5,
    "reason": "Board-approved extension for holiday guests, case #482",
    "createdByAdminID": "A0000001",
    "createdAt": "2026-09-24T18:04:00Z"
  }
]
```

Empty history returns `[]`, not `404` (resident exists, they simply have no adjustments — effective allowance is the standard 20).

## Modified: `GET /api/resident/{id}` and `GET /api/residents`

**Change**: The resident JSON representation no longer includes a writable `amtParkingDaysUsed` field. It is replaced by two read-only, derived fields:

```json
{
  "id": "B1234567",
  "firstName": "...",
  "...": "...",
  "unlimDays": false,
  "daysUsed": 18,
  "effectiveAllowance": 25
}
```

`daysUsed` and `effectiveAllowance` are computed per the formulas in `data-model.md` and appear on every response that currently includes resident data — they are not a separate lookup.

## Modified: `PUT /api/resident` (edit)

**Change**: The request body no longer accepts `amtParkingDaysUsed` as an editable field. Per FR-009, any attempt to include it is ignored by the handler (the field is simply not part of `EditResident`'s accepted fields anymore — same mechanism the existing `password`/`unlimDays` fields already use for "optional, settable if present"). This is a behavior change from today, where `amtParkingDaysUsed` is currently accepted and directly overwrites the stored counter — that is precisely the capability this feature removes.

## Modified: `POST /api/permit`

**Change**: No request/response shape change. The validation performed when the resident is not `unlimDays` and the permit is not an exception now compares `derived days_used + requested permit days` against `derived effective_allowance` instead of the old stored `amtParkingDaysUsed`. The two existing error responses (`errs.EntityDaysTooLong`, `errs.PermitPlusEntityDaysTooLong`) are reused with the derived values substituted in, so client-visible error message shape is unchanged — only the numbers behind them are now guaranteed-correct.
