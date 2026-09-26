# Contract: Unified Quota Ledger

Follows this codebase's existing route conventions exactly: flat, singular-noun paths under `/api`, JSON request/response bodies, errors returned via the existing `errs.APIErr` → `respondError` convention (HTTP status + message), auth via the existing `middleware.authenticate(role...)` chi middleware.

## `POST /api/quota-adjustment`

**Auth**: `AdminRole` only (same group as `POST /api/resident`).

Records one manual (`entry_type = 'ADMIN_ADJUSTMENT'`) ledger entry. This is the *only* way a quota-affecting event is created directly through this endpoint — permit-driven entries are always created as a side effect of `POST`/cancel on `/api/permit` (below), never through this route.

**Request body**:

```json
{
  "residentID": "B1234567",
  "amount": 5,
  "reason": "Board-approved extension for holiday guests, case #482"
}
```

**Behavior**:
1. Reject if `residentID` does not reference an existing resident → `404` (`errs.NewNotFound("resident")`).
2. Reject if `amount == 0` → `400` (new error: "amount must be non-zero").
3. Reject if `reason` is empty/whitespace-only → `400` (new error: "reason must not be empty") — FR-005 / edge case.
4. `createdByAdminID` is always taken from the authenticated admin's access-token payload (`ctxGetAccessPayload`), never from the request body.
5. `quotaYear` is computed server-side at write time (research.md §3) — not accepted from the request body.
6. On success, insert the row and return it.

**Response** (`200`):

```json
{
  "id": 42,
  "residentID": "B1234567",
  "quotaYear": 2026,
  "amount": 5,
  "entryType": "ADMIN_ADJUSTMENT",
  "permitID": null,
  "reason": "Board-approved extension for holiday guests, case #482",
  "createdByAdminID": "A0000001",
  "createdAt": "2026-09-26T18:04:00Z"
}
```

**Not supported**: `PUT`/`PATCH`/`DELETE` on this resource — there is no route to edit or remove a ledger entry (FR-002). Corrections are made by `POST`-ing a new, offsetting adjustment.

## `GET /api/resident/{id}/quota-history`

**Auth**: `AdminRole` only. (Residents do not get entry-level access to their own history — clarification session 2026-09-26 — they only see the two derived totals below.)

**Behavior**: Returns the resident's complete, unified history — both `PERMIT` and `ADMIN_ADJUSTMENT` entries — in a single chronological list (oldest first), satisfying FR-009/User Story 1. This replaces the superseded `GET /api/resident/{id}/quota-adjustments` route, which only ever returned admin-adjustment rows.

**Response** (`200`):

```json
[
  {
    "id": 101,
    "residentID": "B1234567",
    "quotaYear": 2026,
    "amount": 3,
    "entryType": "PERMIT",
    "permitID": 55,
    "reason": null,
    "createdByAdminID": null,
    "createdAt": "2026-09-24T10:00:00Z"
  },
  {
    "id": 102,
    "residentID": "B1234567",
    "quotaYear": 2026,
    "amount": -3,
    "entryType": "PERMIT",
    "permitID": 55,
    "reason": null,
    "createdByAdminID": null,
    "createdAt": "2026-09-25T09:00:00Z"
  },
  {
    "id": 103,
    "residentID": "B1234567",
    "quotaYear": 2026,
    "amount": 5,
    "entryType": "ADMIN_ADJUSTMENT",
    "permitID": null,
    "reason": "Board-approved extension for holiday guests, case #482",
    "createdByAdminID": "A0000001",
    "createdAt": "2026-09-26T18:04:00Z"
  }
]
```

Empty history returns `[]`, not `404` (resident exists, they simply have no history yet — effective allowance is the standard 20, usage is 0).

## Modified: `GET /api/resident/{id}` and `GET /api/residents`

**Change**: The resident JSON representation no longer includes a writable `amtParkingDaysUsed` field. It is replaced by two read-only, derived fields, computed via one grouped query across the whole page (FR-010/FR-015, never a per-resident follow-up query):

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

## Modified: `PUT /api/resident` (edit)

**Change**: The request body no longer accepts `amtParkingDaysUsed` as an editable field (FR-003). Any such field in the request body is ignored by the handler — the field is simply not part of `EditResident`'s accepted fields anymore.

## Modified: `POST /api/permit`

**Change**: No request/response shape change. Creation now runs inside the locked transaction described in research.md §6: the resident's derived usage (current `quotaYear`) plus the requested permit's days is checked against their effective allowance, the permit row is inserted, and a same-transaction `PERMIT`-type ledger entry is inserted. The two existing error responses (`errs.EntityDaysTooLong`, `errs.PermitPlusEntityDaysTooLong`) are reused with the derived values substituted in, so client-visible error shape is unchanged.

## Modified: `DELETE /api/permit/{id}`

**Change**: No longer physically deletes the permit row (FR-016). Instead, in one transaction: sets `cancelled_ts`, and inserts an offsetting `PERMIT`-type ledger entry (`amount = -days`, same `quotaYear` as the original creation entry — FR-017). Response shape is unchanged (`{"message": "Successfully deleted permit"}`); the permit remains fetchable via `GET /api/permit/{id}` afterward (excluded from all list/active-permit-overlap queries, per data-model.md).
