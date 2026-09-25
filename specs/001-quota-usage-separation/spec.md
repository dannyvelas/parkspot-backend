# Feature Specification: Separate Permit Usage from Admin-Editable Allowance

**Feature Branch**: `001-quota-usage-separation`

**Created**: 2026-09-24

**Status**: Draft

**Input**: User description: "this is a very simple REST backend. it basically just allows you to do CRUD operations on the database table \"permits\" and \"residents\" (among other things). there are a few rules that determine whether a resident can create a permit like \"every permit has a count of days, and the count of days of all of a users permits can be no greater than 20\". the count of the amount of days that a resident `r` has used in permits is stored in the \"residents\" table, in the row that corresponds to `r`. the problem is that sometimes the table will get to a state where nobody understands how it got there. For example, suppose a resident `r` created permits that add up to a total of 20 days. suppose that after this, an admin decrement's `r`'s \"day-count\" field to be 17, and then `r` creates a new permit for 3 days. at this point, the table ends up in a state where it's hard to tell whether the application has a bug that allowed the user to create the new permit, or if it was valid because of an admin edit. Because of this: stop storing days used as a single writable number — derive it from permits, and hold admin-set changes to the limit in a separate, auditable adjustment record. Validation becomes: days used <= base limit (20) + sum of adjustments."

## Clarifications

### Session 2026-09-24

- Q: Does days-used include permits whose date range has already passed (expired), or only permits currently in effect? → A: Days-used includes ALL of a resident's non-deleted permits, regardless of whether their date range has passed (lifetime running total, matches current behavior).
- Q: The system also tracks a similar day-usage counter per car (incremented/decremented alongside the resident counter on every permit create/delete). Is fixing that counter in scope for this feature? → A: Out of scope — this feature only fixes the resident-level days-used/allowance split; the car-level counter is untouched and may be addressed separately.
- Q: If an admin records an allowance adjustment by mistake (e.g., a typo), can that adjustment be edited or deleted afterward? → A: No — the ledger is append-only. Mistakes are corrected by recording a new offsetting adjustment; the original entry is never edited or deleted.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Trustworthy, Always-Accurate Usage Tracking (Priority: P1)

As a property administrator, when I look up how many parking days a resident has used, I want that number to always reflect their actual current permits — so I never have to guess whether it's correct or was altered by hand.

**Why this priority**: This is the core problem the feature exists to solve. Today the "days used" number can be edited directly, which is exactly what makes it untrustworthy. Nothing else in this feature matters if this isn't fixed first.

**Independent Test**: Create several permits for a resident totaling a known number of days and confirm the displayed usage matches the sum exactly; delete one permit and confirm usage decreases by that permit's day count automatically, with no manual step.

**Acceptance Scenarios**:

1. **Given** a resident with no permits, **When** an admin views the resident's record, **Then** the displayed days-used value is 0.
2. **Given** a resident who creates permits totaling 12 days, **When** an admin views the resident's record, **Then** the displayed days-used value is exactly 12.
3. **Given** a resident with permits totaling 12 days, **When** one 4-day permit is deleted, **Then** the displayed days-used value automatically becomes 8, with no admin editing the value directly.
4. **Given** the system in any state, **When** any user, including an administrator, attempts to directly set or overwrite the days-used value, **Then** the system rejects the change and indicates that usage is calculated automatically from permits.

---

### User Story 2 - Explicit, Auditable Allowance Adjustments (Priority: P1)

As a property administrator, when a resident needs extra parking days beyond the standard limit (e.g., a documented special circumstance), I want to record that as a clear, traceable grant to their allowance — not as an edit to their usage — so anyone reviewing the account later can see exactly why the resident was able to exceed the standard limit.

**Why this priority**: This is the other half of the core fix. It replaces the ambiguous historical pattern (admin edits usage) with an unambiguous one (admin edits allowance, on the record), which is the specific mechanism the requester asked for.

**Independent Test**: As an admin, add a +5 day allowance adjustment to a resident with a documented reason; confirm the resident's effective limit becomes 25 while their days-used value is unaffected; confirm the adjustment appears in that resident's adjustment history with who made it, when, why, and the amount.

**Acceptance Scenarios**:

1. **Given** a resident at the standard 20-day limit, **When** an admin records a +5 day allowance adjustment with a reason, **Then** the resident's effective limit becomes 25 and their days-used value is unchanged.
2. **Given** a resident with an allowance adjustment on file, **When** an admin reviews that resident's account, **Then** they can see the adjustment's amount, reason, responsible admin, and timestamp.
3. **Given** a resident with multiple allowance adjustments over time, **When** an admin reviews the resident's account, **Then** all adjustments are shown individually (none overwritten or lost), and the effective limit reflects the standard base plus the sum of all adjustments.

---

### User Story 3 - Permit Creation Validated Against Real Usage and Allowance (Priority: P2)

As a resident, when I request a new permit, I want the system to accept or reject it based on my actual current usage and my true allowance (standard limit plus any admin grants) — so the outcome is predictable and consistent with what an admin can see on my account.

**Why this priority**: This ensures the fix isn't just cosmetic — the gating logic itself uses the new, trustworthy values, so every accepted or rejected permit request is explainable after the fact.

**Independent Test**: With a resident at 18 of 20 days, attempt to create a 3-day permit (should be denied, exceeds limit); add a +5 day admin adjustment; retry the same 3-day permit request (should now succeed, since 18+3=21 is within the new 25-day allowance).

**Acceptance Scenarios**:

1. **Given** a resident whose current usage plus a requested permit's days would exceed their effective limit (base plus adjustments), **When** they submit the permit request, **Then** the system rejects it and explains that it would exceed their allowed parking days.
2. **Given** a resident whose current usage plus a requested permit's days is within their effective limit, **When** they submit the permit request, **Then** the system accepts it.
3. **Given** a resident who is flagged for unlimited parking days, **When** they submit any permit request, **Then** the day-limit check does not apply.

---

### User Story 4 - Historical Adjustment Visibility for Investigations (Priority: P3)

As a property administrator investigating a resident's account (e.g., "why does this resident have 23 used days when the limit is 20?"), I want to see a complete, chronological record of every allowance adjustment ever made to that resident, so I can answer the question without guessing or escalating to engineering.

**Why this priority**: This directly resolves the original pain point described in the request — not being able to tell whether an above-limit state reflects a bug or a legitimate admin decision. It's high value, but the system is still usable without it (totals from User Story 2 already show the current effective limit), so it's P3 rather than P1.

**Independent Test**: With a resident who has received two separate allowance adjustments at different times, request the resident's adjustment history and confirm both appear with correct amounts, timestamps, reasons, and responsible admin, in chronological order.

**Acceptance Scenarios**:

1. **Given** a resident with two allowance adjustments made on different dates, **When** an admin requests that resident's adjustment history, **Then** both adjustments are listed with amount, reason, responsible admin, and timestamp, in chronological order.
2. **Given** a resident with no allowance adjustments, **When** an admin requests that resident's adjustment history, **Then** the history is empty and the resident's effective limit equals the standard base limit.

---

### Edge Cases

- A resident whose recorded days-used value already exceeds their standard base limit at the moment this change takes effect (from a past direct edit with no equivalent adjustment on record) receives an automatically generated migration-reconciliation adjustment that preserves their prior effective limit, so they are not newly blocked from anything they could do immediately beforehand.
- Permits explicitly marked as exempt from the day limit (e.g., an approved exception) must not count toward days-used.
- Residents flagged for unlimited parking days are excluded from the day-limit check entirely — their days-used and effective allowance are not evaluated for permit creation.
- If two permit requests for the same resident are submitted at nearly the same time, the system must not allow both to succeed if only one would fit within the resident's effective allowance.
- An administrator attempting to record an allowance adjustment without a reason must be blocked, since every adjustment must be explainable later.
- An administrator attempting to edit or delete an existing allowance adjustment (e.g., to fix a mistake) must be blocked; the correction must instead be a new offsetting adjustment.
- Deleting a permit that previously counted toward days-used must automatically decrease that resident's days-used value, with no separate admin action required.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST calculate each resident's days-used value solely as the sum of days across that resident's non-deleted permits that count toward their limit (excluding permits explicitly exempted from the limit and excluding residents flagged for unlimited days). This sum includes permits whose date range has already passed — an expired permit continues to count until it is deleted; days-used is not reduced by the mere passage of time. This value MUST NOT be directly settable by any user, including administrators.
- **FR-002**: System MUST update a resident's days-used value automatically and immediately whenever a qualifying permit is created or deleted, with no manual recalculation step required by any user.
- **FR-003**: System MUST allow administrators to record an allowance adjustment for a resident, capturing at minimum: the adjustment amount, a required reason, the identity of the administrator who made it, and the date/time it was made.
- **FR-004**: System MUST treat allowance adjustments as an append-only, individually-retained ledger — once created, an adjustment MUST NOT be edited or deleted by any user, including administrators. A mistaken adjustment MUST be corrected by recording a new, separate offsetting adjustment (e.g., a −45 entry to counteract an erroneous +50 entry), never by altering the original.
- **FR-005**: System MUST compute each resident's effective allowance as the standard base limit (20 days) plus the sum of all of that resident's allowance adjustments.
- **FR-006**: System MUST reject a new permit request when the resident's days-used plus the requested permit's day count would exceed their effective allowance, unless the resident is flagged for unlimited days or the permit is explicitly exempt from the limit.
- **FR-007**: System MUST allow administrators to view, for any resident, the full history of that resident's allowance adjustments (amount, reason, responsible admin, timestamp), in chronological order.
- **FR-008**: System MUST present a resident's current days-used value and current effective allowance as two distinct figures, so usage and allowance are never displayed as a single ambiguous number.
- **FR-009**: System MUST prevent any user role, including administrators, from directly editing or overwriting the days-used value; the only way that value changes is by creating or deleting a qualifying permit.
- **FR-010**: System MUST reconcile residents whose historical days-used value does not match the sum of their qualifying permits when this change takes effect, by automatically generating a one-time allowance adjustment for each such resident that preserves their currently-effective limit (i.e., the difference between their prior recorded usage and their true derived usage becomes an adjustment, with a reason indicating it is a migration reconciliation). No resident MUST become newly blocked from an action they could take immediately before this change.

### Key Entities

- **Resident**: The person requesting parking permits. Relevant attributes: whether they are flagged for unlimited parking days (exempt from the limit entirely), a derived days-used value (never directly editable), and a computed effective allowance (standard base limit plus their adjustments).
- **Permit**: A request for parking privileges covering a specific date range for a specific vehicle. Relevant attributes: the resident it belongs to, the number of days it spans, and whether it counts toward that resident's day limit (some permits are explicitly exempted).
- **Allowance Adjustment**: An explicit, auditable record of an administrator changing a resident's allowance, not their usage. Relevant attributes: the resident it applies to, the adjustment amount, the reason, the responsible administrator, and when it was made.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Given the same account state, any two administrators reviewing a resident's account independently reach the same conclusion about whether that resident's current parking day usage is valid, without consulting anyone else.
- **SC-002**: 100% of changes to a resident's effective parking-day allowance are traceable to a specific admin, reason, and timestamp, with zero allowance changes that cannot be explained by an on-record adjustment.
- **SC-003**: A resident's displayed days-used value matches the sum of their currently-qualifying permits' days in 100% of cases, with zero manual reconciliation required by staff.
- **SC-004**: Every "how did this resident's usage get to this number" investigation is resolvable by an administrator using only the adjustment history and current permits, without engineering involvement.
- **SC-005**: No permit is ever accepted that pushes a resident's usage over their effective allowance, except where the resident is flagged for unlimited days or the permit is explicitly exempt.

## Out of Scope

- The per-car day-usage counter (a separate, similar running total kept per vehicle) is not changed by this feature. It continues to be maintained as it is today; deriving it from permits and adding car-level allowance adjustments would be a separate effort.

## Assumptions

- The standard base parking-day limit remains a fixed, system-wide value of 20 days, applied uniformly to all residents except those flagged for unlimited days, consistent with current behavior.
- Allowance adjustments may be positive (grants) or negative (reductions); every adjustment requires a reason regardless of direction so the record stays meaningful.
- "Permits that count toward the limit" follows existing behavior: permits explicitly marked as exempt (e.g., an approved exception) and permits belonging to a resident flagged for unlimited days do not contribute to days-used.
- Only administrators can create allowance adjustments; residents can view their own current days-used and effective allowance but do not need direct access to the adjustment history for this feature to deliver value.
- This feature does not introduce a time-based reset (e.g., annual renewal) of the day limit; the limit remains a running total unless changed by a future feature.
- At migration time, any resident whose previously recorded days-used value differs from the true sum of their qualifying permits receives a one-time, system-generated allowance adjustment (reason: migration reconciliation) that preserves their currently-effective limit, so no resident is newly blocked from an action they could take immediately before the change.
