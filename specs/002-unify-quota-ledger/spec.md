# Feature Specification: Unify Permit and Admin Quota Changes into One Ledger

**Feature Branch**: `002-unify-quota-ledger`

**Created**: 2026-09-26

**Status**: Draft

**Input**: User description: "i've been running speckit in this repo to fix an auditability problem in the database this repo uses. i implemented the first 3 tasks before realizing that actually the solution i was about to implement was a bit suboptimal. the solution i was about to implement: had an N+1 problem (whenever i would query a list of residents i would have to make 2 database queries for each resident to determine the count of days that they had left); it made the ledger only take into adjustments to resident days instead of being the single source-of-truth for how many days a resident should have (this table was called quota_adjustments, so calculating a resident's days required querying both the permits table and the quota_adjustments table). Now I'd like to instead: stop treating days-used as a stored fact, and derive it from a single append-only ledger table that is the ONLY source of truth for quota-affecting events, including permits (not just admin adjustments). Ledger rows are never updated or deleted — cancelling a permit inserts a reversing entry rather than deleting the original. Creating or cancelling a permit writes a corresponding ledger row in the same transaction. Listing residents' balances becomes a single grouped query instead of N+1 queries. The annual reset disappears entirely: a new allowance period's balance is naturally zero since no rows exist yet for that period. Carryover of unused quota is explicitly out of scope for v1. This supersedes the existing quota_adjustments-based spec/plan/tasks (001-quota-usage-separation); tasks 1-3 of that plan were already implemented and merged, tasks 4-5 were never merged and should not be carried forward as-is."

## Clarifications

### Session 2026-09-26

- Q: Should a cancelled permit's own record be preserved (soft-cancel) so ledger entries stay traceable to a viewable permit, or can the permit row be hard-deleted as it is today? → A: Permits are never hard-deleted going forward; cancelling a permit marks it as cancelled and its record remains stored and queryable, so every ledger entry's referenced permit stays permanently inspectable.
- Q: When a permit is cancelled, does its reversing entry count against the same allowance period as the original permit entry, or the period current at the moment of cancellation? → A: The same period as the original entry, always. Combined with period-at-creation attribution (FR-011), this means a permit created near the end of one period for dates in the next period is booked entirely against the *creation* period; cancelling it later nets that same, possibly already-closed, period back to zero and leaves every other period — including whichever period the permit's own active dates fell in — completely untouched.
- Q: At cutover, should every resident's existing (pre-cutover) permits each get their own individual history entry, or should all pre-cutover activity be collapsed into one lump reconciling entry per resident? → A: Individual entries — every existing qualifying permit is backfilled with its own history entry (attributed to that permit's original creation time and period), so pre-cutover activity is exactly as auditable, permit-by-permit, as anything created afterward. Only a genuine residual mismatch (e.g., leftover drift not explained by any existing permit or already-recorded admin adjustment) becomes a single final reconciling entry per resident, rather than the whole history being one lump sum.
- Q: Once existing `quota_adjustments` rows are folded into the new unified history, should that table be removed, or kept as a secondary historical artifact? → A: Removed — it serves no purpose once cutover is complete, and its continued presence would recreate the "two sources of truth" ambiguity this feature exists to end. In practice the table is currently empty (never used in production), so cutover has no `quota_adjustments` rows to carry forward; any residual mismatch this feature reconciles comes from drift between the old, superseded stored usage number and residents' true permit history, not from that table.
- Q: Should residents be able to see their own entry-level history, or only their current usage/allowance totals? → A: Totals only — entry-level history detail (which permit, which admin, which reason) remains admin-only, matching the superseded spec's boundary; this feature's scope stays focused on the data model and admin-facing auditability problem, not new resident-facing history UI/API.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - One Unified, Unchangeable History for Every Quota Change (Priority: P1)

As a property administrator investigating a resident's current parking-day usage, I want every event that ever changed that number — whether it came from a resident creating or cancelling a permit, or from an admin recording a manual adjustment — recorded in a single chronological, tamper-proof history, so I can always explain exactly how the current number was reached without cross-referencing separate records or guessing which one is authoritative.

**Why this priority**: This is the core problem the feature exists to solve. Splitting the audit trail between permits and a separate adjustments record (the previous approach) still leaves two sources of truth that can drift apart or require separate lookups to reconcile. Nothing else in this feature matters if a single, complete history isn't achieved first.

**Independent Test**: Create a permit for a resident (usage increases by its day count), cancel that permit (usage decreases, via a new offsetting entry rather than erasing the original), and have an admin record a manual adjustment with a reason. Confirm all three events appear for that resident, in order, each clearly attributed to its real cause (the specific permit, or the specific admin and reason), with the original permit-creation record still visible and unaltered after the cancellation.

**Acceptance Scenarios**:

1. **Given** a resident with no history, **When** they are issued a 3-day permit, **Then** their current usage is 3 and a single history entry exists showing that permit as the cause.
2. **Given** a resident whose usage includes a 3-day permit, **When** that permit is cancelled, **Then** their current usage returns to what it was before that permit, a new offsetting entry is recorded for the cancellation, and the original entry for the permit's creation remains visible in the history exactly as it was first recorded.
3. **Given** a resident's current usage, **When** an administrator records a manual adjustment with a reason, **Then** the adjustment appears in the resident's history with its amount, reason, responsible administrator, and timestamp, alongside any permit-driven entries, in a single combined timeline.
4. **Given** any entry already recorded in a resident's history, **When** any user, including an administrator, attempts to edit or remove that entry, **Then** the system rejects the change; the only way to correct a mistake is to record a new, separate offsetting entry.
5. **Given** the system in any state, **When** any user attempts to directly set or overwrite a resident's current usage or allowance number, **Then** the system rejects the change, since that number only ever exists as computed from the history.

---

### User Story 2 - Fast, Consistent Usage Totals When Reviewing Many Residents (Priority: P1)

As a property administrator viewing a list of residents, I want to see every resident's current usage without the page getting slower as the number of residents grows, so reviewing the community stays practical regardless of its size.

**Why this priority**: The previous approach required extra lookups per resident shown, which meant a page listing residents got proportionally slower as the resident count grew. That made routine review impractical at scale and is one of the two concrete problems this feature must fix, so it's tied for the highest priority alongside unifying the audit trail.

**Independent Test**: Populate a large number of residents with varied permit and adjustment histories. Confirm that displaying a full page of residents' current usage values requires a bounded amount of work that does not grow per resident shown, and that the displayed totals still exactly match what User Story 1 would show for each resident individually.

**Acceptance Scenarios**:

1. **Given** a page of many residents each with their own permits and adjustments, **When** an administrator opens the resident list, **Then** every resident's current usage is displayed correctly and the amount of work needed to produce the page does not scale with the number of residents shown.
2. **Given** a resident list is displayed, **When** the underlying history for one resident changes (a permit is created, cancelled, or an adjustment is recorded), **Then** the next time the list is viewed it reflects the new total with no manual recalculation or background job required.

---

### User Story 3 - Permit Requests Judged Against the Same Trustworthy Total (Priority: P2)

As a resident requesting a new permit, I want my request evaluated against the same authoritative usage and allowance numbers an administrator would see, so the outcome is predictable and consistent with what's on my account.

**Why this priority**: This ensures the unified history isn't just a display feature — the gating logic itself must use it, so every accepted or rejected permit request is explainable after the fact using the same history from User Story 1.

**Independent Test**: With a resident at 18 of 20 days for the current allowance period, attempt to create a 3-day permit (should be denied). Have an admin record a +5 day adjustment, then retry the same request (should now succeed). Confirm two residents submitting requests for the same resident at nearly the same time cannot both succeed if only one would fit.

**Acceptance Scenarios**:

1. **Given** a resident whose current usage plus a requested permit's days would exceed their effective allowance for the current period, **When** they submit the permit request, **Then** the system rejects it and explains that it would exceed their allowed parking days.
2. **Given** a resident whose current usage plus a requested permit's days is within their effective allowance, **When** they submit the permit request, **Then** the system accepts it and a corresponding history entry is recorded in the same action.
3. **Given** two permit requests for the same resident submitted at nearly the same time, **When** only one would fit within the resident's effective allowance, **Then** the system allows at most one of them to succeed.
4. **Given** a resident flagged for unlimited parking days, **When** they submit any permit request, **Then** the day-limit check does not apply.

---

### User Story 4 - Automatic Renewal at the Start of Each Allowance Period (Priority: P3)

As a property administrator, I want each resident's usage to automatically start fresh at the beginning of each new allowance period, with no manual step for anyone to remember, so residents are never blocked at the start of a new period due to a forgotten administrative task, and the renewal can never be missed, run early, run late, or run twice.

**Why this priority**: This is a real operational risk today (a manual step performed once a year, with no safety net), but the system remains functional without it in the interim since the reset is only needed once a year. It's ranked below the auditability and performance fixes (which affect every single day of usage) and below the write-path consistency in User Story 3.

**Independent Test**: Confirm that at the moment a new allowance period begins, every resident's current usage reads as zero (before any new activity), with no batch job, script, or manual action performed, and confirm that entries from the previous period remain fully intact and visible in history.

**Acceptance Scenarios**:

1. **Given** a resident with a fully-used allowance in the current period, **When** the next allowance period begins, **Then** their usage for the new period reads as zero without any manual or scheduled reset action being performed.
2. **Given** a resident's history from a prior allowance period, **When** the next period begins and new activity occurs, **Then** the prior period's entries remain unchanged and are still viewable, separate from the new period's running total.
3. **Given** a permit that spans the boundary between two allowance periods, **When** its day count is applied, **Then** it is attributed entirely to the allowance period in effect at the time the permit was created, not split across periods.

---

### Edge Cases

- A resident's usage happens to differ from the sum of their currently-qualifying permits and adjustments at the moment this feature takes effect (carried over from before this change): the cutover process backfills an individual entry per existing permit and resolves any remaining, unexplained difference with a single final reconciling entry per resident, without newly blocking any resident from an action they could take immediately beforehand.
- An administrator attempts to record a manual adjustment without a reason: rejected, since every entry not caused by a permit must be explainable.
- A permit is cancelled more than once, or a cancellation is attempted on a permit that was never actually counted (e.g., it was exempt): the system must not record a duplicate or incorrect offsetting entry.
- Permits explicitly marked as exempt from the day limit, and residents flagged for unlimited parking days, are excluded from both the running total and the limit check, exactly as before.
- Two permit requests, or a permit request and a manual adjustment, submitted for the same resident at nearly the same time must not be allowed to both apply if doing so would let the resident's usage silently exceed their effective allowance.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST record every event that changes a resident's parking-day usage — whether generated automatically by creating or cancelling a permit, or recorded manually by an administrator — as an entry in one single, unified history per resident. No other stored number may serve as an independent source of truth for usage.
- **FR-002**: System MUST treat history entries as permanent once created: no entry may be edited or deleted by any user, including administrators. Cancelling a permit MUST be recorded as a new offsetting entry, never as removal of the original entry.
- **FR-003**: System MUST NOT allow any user, including administrators, to directly set or overwrite a resident's current usage or effective allowance; those values only ever exist as the result of totaling that resident's history.
- **FR-004**: System MUST update a resident's current usage automatically and immediately whenever a qualifying permit is created or cancelled, with no separate manual or background step.
- **FR-005**: System MUST allow administrators to record a manual adjustment for a resident, capturing at minimum: the amount, a required reason, the identity of the responsible administrator, and when it was made.
- **FR-006**: System MUST compute each resident's effective allowance for the current allowance period as the standard base limit (20 days) plus the sum of that resident's manual adjustments recorded for that period.
- **FR-007**: System MUST reject a new permit request when the resident's current usage plus the requested permit's day count would exceed their effective allowance for the current period, unless the resident is flagged for unlimited days or the permit is explicitly exempt from the limit.
- **FR-008**: System MUST prevent two concurrent requests affecting the same resident's usage from both succeeding when only one would fit within the resident's effective allowance.
- **FR-009**: System MUST allow administrators to view, for any resident, the complete history of usage-affecting entries (permit-driven and manual), each showing its cause, amount, responsible party (system/permit or specific admin), reason (for manual entries), and timestamp, in chronological order.
- **FR-010**: System MUST support displaying current usage for a list of residents such that the amount of work needed does not grow individually per resident shown (i.e., no per-resident follow-up lookups are required to produce the list).
- **FR-011**: System MUST attribute each qualifying permit's full day count to the single allowance period in effect at the time the permit was created, even when the permit's own date range extends into a later period.
- **FR-012**: System MUST cause each new allowance period to begin with zero recorded usage for every resident, without requiring any manual, scheduled, or batch action to bring this about.
- **FR-013**: System MUST, at the point this feature takes effect, backfill an individual history entry for every resident's existing qualifying permit, each attributed to that permit's own original creation time and allowance period, so pre-cutover activity is exactly as individually auditable as anything created afterward. Any remaining difference between a resident's previously recorded usage and the total produced by this backfill MUST be resolved by recording one final reconciling entry per resident that preserves their currently-effective allowance, so no resident becomes newly blocked from an action they could take immediately beforehand.
- **FR-014**: System MUST remove the superseded admin-adjustments table (and its associated model/code) once this feature is in place; it MUST NOT be kept alongside the new unified history as a second, parallel record of admin-driven changes.
- **FR-015**: System MUST present a resident's current usage and current effective allowance as two distinct figures, never combined into one ambiguous number.
- **FR-016**: System MUST NOT delete a permit's own record when it is cancelled; the permit MUST remain stored with a cancelled status so that any history entry referencing it stays permanently viewable back to that specific permit.
- **FR-017**: System MUST attribute a permit cancellation's offsetting entry to the same allowance period as that permit's original entry (per FR-011, the period in effect when the permit was created), never to whichever period happens to be current at the moment of cancellation. A cancellation therefore never changes any period's total other than the one the original permit was booked against.

### Key Entities

- **Resident**: The person requesting parking permits. Holds no directly stored usage number; whether they are flagged for unlimited parking days; both current usage and effective allowance are always derived from their history.
- **Permit**: A request for parking privileges covering a specific date range for a specific vehicle. Attributes: the resident it belongs to, the number of days it spans, whether it counts toward that resident's limit, and whether it has been cancelled. A permit is never removed from storage; cancelling it changes its status but keeps it viewable so any history entry it caused remains traceable back to it. Creating or cancelling a qualifying permit always produces a corresponding history entry.
- **Quota History Entry**: A single, permanent record of one event that changed a resident's usage — either generated automatically alongside a permit's creation or cancellation, or recorded manually by an administrator with a reason. Attributes: the resident it applies to, the signed amount, which allowance period it counts against, its cause (the specific permit, or the specific responsible administrator plus reason), and when it was recorded.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Given the same account history, any two administrators reviewing a resident's account independently reach the same conclusion about how that resident's current usage arose, without consulting anyone else or any other record.
- **SC-002**: 100% of changes to a resident's usage, whether caused by a permit or a manual adjustment, are individually traceable to a specific cause and timestamp, with zero changes that cannot be explained by an entry in that resident's history.
- **SC-003**: The time needed to display current usage for a full page of residents does not measurably increase as the number of residents shown grows (e.g., a page of 500 residents loads no slower, in the number of lookups required, than a page of 20).
- **SC-004**: Every "how did this resident's usage get to this number" investigation is resolvable by an administrator using only that resident's history, without engineering involvement.
- **SC-005**: No permit is ever accepted that pushes a resident's usage over their effective allowance for the current period, except where the resident is flagged for unlimited days or the permit is explicitly exempt.
- **SC-006**: At the start of each new allowance period, 100% of residents show zero usage for that period with zero manual steps performed by staff.

## Out of Scope

- Carrying over unused allowance from one period into the next. Every resident starts each new period at the standard base limit with no rollover; this design does not preclude adding carryover later, but v1 does not implement it.
- The per-car day-usage counter (a separate, similar running total kept per vehicle) is not changed by this feature. It continues to be maintained as it is today.
- Allowance periods other than a calendar year (e.g., resident-specific anniversary-based periods).

## Assumptions

- The standard base parking-day limit remains a fixed, system-wide value of 20 days per allowance period, applied uniformly to all residents except those flagged for unlimited days.
- The allowance period is the calendar year, matching current expectations for an annual reset.
- Only administrators can record manual (non-permit) history entries. Residents can view their own current usage and effective allowance but not entry-level history detail — confirmed during clarification, matching the superseded spec's boundary.
- "Permits that count toward the limit" follows existing behavior: permits explicitly marked as exempt, and permits belonging to a resident flagged for unlimited days, do not contribute to usage.
- This feature supersedes the in-progress `001-quota-usage-separation` effort. The already-merged portions of that effort (the separate admin-adjustments table, its migration backfill, and its model) are removed and superseded by this feature's unified history rather than extended further; that table currently holds no rows (it was never used in production), so removing it carries forward no data. The unmerged portions of that effort (its remaining tasks, present only in unmerged branches) are not carried forward as-is, since they were designed around a two-table split this feature eliminates.
- At the point this feature takes effect, every resident's existing permits are individually backfilled into the new unified history (see FR-013), and any remaining unexplained difference is resolved with one final reconciling entry per resident, so that every resident's effective allowance immediately beforehand is preserved.
