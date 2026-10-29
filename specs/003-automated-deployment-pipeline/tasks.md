---

description: "Task list for feature implementation"
---

# Tasks: Automated CI/CD Deployment Pipeline

**Input**: Design documents from `/specs/003-automated-deployment-pipeline/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md (all present)

**Tests**: No automated test tasks are generated for the pipeline/ops configuration itself (not requested in the spec, and GitHub Actions/Kamal/cron config isn't unit-testable in the Go suite). Each user story instead ends with a task to run its corresponding scenario in `quickstart.md`, which is this feature's acceptance test. The one piece of actual Go application code this feature adds (`GET /up`) does get a Go test file, matching the existing per-handler test convention in `api/`.

**Organization**: Tasks are grouped by user story (from `spec.md`) to enable independent implementation and validation of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1–US6)
- Exact file paths are included in each description

## Path Conventions

Per `plan.md`'s Project Structure: application code changes live in `api/`; everything else is new pipeline/ops configuration at the repo root (`Dockerfile`, `config/deploy.yml`, `.kamal/`, `.github/workflows/`, `ops/`).

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Repo-level scaffolding needed before any pipeline plumbing can be wired up.

- [X] T001 [P] Enable GitHub branch protection on `main` (repo Settings → Branches): require a Go test-suite status check to pass, disallow direct pushes, and note that admins may still bypass via force-merge (per FR-003/FR-004 — the deploy pipeline will trust this and not re-check). **Prerequisites discovered during implementation**: (1) this repo had zero GitHub Actions workflows configured, so there was no existing status check to require — added `.github/workflows/test.yml` (`go build`/`go test` on PRs into `main`); (2) that workflow initially failed in CI — `api` package tests need `OAUTH_*` env vars (fixed by copying the repo's own non-secret `.env.example` to `.env` in the workflow) and `app` package tests hit a pre-existing testcontainers Postgres race (fixed the wait strategy in `storage/psql/test_helpers.go`). Branch protection applied via `gh api PUT .../branches/main/protection`: required status check `test`, `strict: true`, `enforce_admins: false` (admins can still force-merge per FR-004), `allow_force_pushes`/`allow_deletions: false`.
- [X] T002 [P] Create directory skeletons: `.github/workflows/`, `.kamal/hooks/`, `ops/backup/` (empty `.gitkeep` or first real file in each, per `plan.md` Project Structure).
- [X] T003 [P] Add `Dockerfile` at repo root: multi-stage build producing a minimal production image of the existing Go binary (`go build ./...` stage on a Go 1.25 base image, copied into a slim runtime base image), exposing `$PORT`.
- [X] T004 [P] Add `.dockerignore` at repo root excluding `.git`, `.db/`, local `.env`, and other non-build files from the Docker build context.

**Checkpoint**: Repo scaffolding exists; branch protection is enforced.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core plumbing that every user story depends on — without this, no deploy of any kind can happen to validate any story.

**⚠️ CRITICAL**: No user story work can be meaningfully validated until this phase is complete.

- [X] T005 [P] Configure the one-time CI/host credentials as GitHub Actions encrypted secrets (repo or environment settings): SSH deploy key for the production host, Bitwarden Secrets Manager access token (`bws`), and the Slack/Discord webhook URL. `ghcr.io` auth uses the workflow's built-in `GITHUB_TOKEN` (no extra secret needed), per `research.md` Decision 4. Confirmed present via `gh secret list`: `SSH_PRIVATE_KEY`, `BWS_ACCESS_TOKEN`, `DEPLOY_WEBHOOK_URL`.
- [X] T006 [P] Create `config/deploy.yml` (Kamal config): app service pointing at the `Dockerfile` from T003, the production server's SSH target, `ghcr.io` as the registry, and a Postgres accessory (pinned image tag, named persistent volume) per `data-model.md`'s Server Environment entity. Server target: `deploy@park-spot.co` (provided by operator). Postgres pinned to `17.6` as a default — **confirm/adjust this if an existing production database is already running a different major version** before T008's first bootstrap. `proxy.healthcheck` is intentionally left at Kamal's default (checks `/`, which the app already serves) since the dedicated `/up` endpoint doesn't exist until US2 (T018 switches it over).
- [X] T007 [P] Create `.kamal/secrets`: resolves application env vars (`DATABASE_URL`, `TOKEN_ACCESSSECRET`, `TOKEN_REFRESHSECRET`, `OAUTH_CLIENTID`, `OAUTH_CLIENTSECRET`, `OAUTH_REDIRECTURL`, `OAUTH_SCOPE`, `OAUTH_AUTHURL`, `OAUTH_TOKENURL`, `OAUTH_ACCESSTOKEN`, `OAUTH_REFRESHTOKEN`, `OAUTH_TOKENTYPE`, `OAUTH_EXPIRY`) from the GitHub Actions secrets set up in T005 and/or `bws` CLI calls, per `data-model.md`'s Secret entity. No value is hardcoded in this file. Uses Bitwarden Secrets Manager project `c902bd7c-b27c-4791-bab1-b4e5003dac58` via `bws secret list ... --output env` — every secret listed in the file (including `POSTGRES_PASSWORD` for the accessory) must exist under that exact name in that project.
- [X] T008 Run `kamal setup` once against the real production server, using `config/deploy.yml` from T006 and `.kamal/secrets` from T007, to perform the very first bootstrap (installs Docker, boots the app + Postgres accessory). This is a one-time manual operation that makes every subsequent story testable; it is superseded as a *repeatable* process by T029 in US5. Done — the server has been running the app, Postgres accessory, and (separately) the frontend via Kamal for several days, including a full production data migration/reconciliation on top of it.
- [X] T009 [P] Confirm the branch-protection required check configured in T001 is actually the repo's existing Go test-suite CI job (not a stale/misnamed check), by opening a throwaway PR and observing it run. Already confirmed organically while landing T001 itself: PRs #37 and #38 both ran and were gated on the real `test` check (`go build ./...` + `go test ./...`), including one case where it genuinely failed and blocked merge until fixed — no separate throwaway PR needed.

**Checkpoint**: Foundation ready — a manually-triggered `kamal deploy` against `config/deploy.yml` now works end-to-end. User story implementation can begin.

---

## Phase 3: User Story 1 - Merge to main automatically ships to production (Priority: P1) 🎯 MVP

**Goal**: A merged, CI-passed pull request into `main` results in production running that code automatically, with no manual `scp`/build step, and a direct push to `main` never triggers a deploy.

**Independent Test**: Merge a trivial, approved PR into `main` and confirm the running production service reflects the new code without any manual deployment action.

### Implementation for User Story 1

- [X] T010 [US1] Create `.github/workflows/deploy.yml`: triggers on `pull_request` `closed` events (or `push` to `main`, gated — see T011), checks out the merge commit, builds and pushes the Docker image from T003 to `ghcr.io`, then runs `kamal deploy` using `config/deploy.yml` (T006) and `.kamal/secrets` (T007). Build+push is handled implicitly by `kamal deploy` itself (Kamal builds from the Dockerfile and pushes to the registry configured in `config/deploy.yml` as one step), not a separate action. Checks out `ref: main` explicitly — by the time a `pull_request: closed` (merged) event fires, `main`'s tip already is the squash-merge commit, so no special PR-merge-preview ref is needed. **Known gap until T011 lands**: this trigger (`pull_request: types: [closed]`) fires on *any* closed PR, merged or not — do not merge this to `main` without T011's guard landing in the same PR.
- [ ] T011 [US1] In `.github/workflows/deploy.yml`, add the job-level guard from `contracts/ci-deploy-trigger.md` so the deploy job only runs when `github.event.pull_request.merged == true` and the base ref is `main` — explicitly excluding closed-but-not-merged PRs and any push to `main` that isn't the result of a merged PR.
- [ ] T012 [US1] In `.github/workflows/deploy.yml`, write a `$GITHUB_STEP_SUMMARY` (or equivalent) recording the Deployment record from `data-model.md`: commit SHA, PR number, image tag, and final status — so deployment history is reviewable after the fact (FR-020).
- [ ] T013 [US1] Confirm (and comment inline in `.github/workflows/deploy.yml`) that the workflow has no test-running step of any kind, per FR-004 — merged code is trusted as already validated by the branch-protection check from T001.
- [ ] T014 [US1] Run `quickstart.md` Scenario 1 end-to-end: merge a trivial PR and confirm the automatic deploy; separately confirm a direct push to `main` triggers no deploy.

**Checkpoint**: At this point, merging a PR into `main` ships it to production automatically. This is the MVP.

---

## Phase 4: User Story 2 - Safe deploys with automatic health validation and rollback (Priority: P1)

**Goal**: Every deploy is gated on a real health check; a bad deploy never gets production traffic and production automatically stays on the last known-good version, with zero end-user-visible downtime.

**Independent Test**: Deploy a version that deliberately fails its health endpoint and confirm production automatically keeps serving the previous version with no downtime and no operator action.

### Implementation for User Story 2

- [ ] T015 [US2] Implement `GET /up` in `api/health_handler.go` per `contracts/health-check-endpoint.md`: unauthenticated, returns `200 ok` only once the process confirms it can reach Postgres via the existing `storage/psql` connection; returns `503` otherwise. No version/internal info is exposed.
- [ ] T016 [US2] Wire the `/up` route into the existing unauthenticated route group in `api/server.go` (alongside the current `/` "hello world" route), depends on T015.
- [ ] T017 [US2] Add `api/health_handler_test.go` covering the healthy and DB-unreachable cases, matching the existing per-handler test convention (e.g. `car_handler_test.go`), depends on T015.
- [ ] T018 [US2] Configure the `proxy.healthcheck` block in `config/deploy.yml` (from T006) to poll `GET /up` with a bounded number of attempts/interval before kamal-proxy cuts traffic over to a new container, depends on T006, T015.
- [ ] T019 [US2] Verify the existing graceful-shutdown window in `main.go` (`server.ShutdownGracefully(30 * time.Second)`) is long enough to drain in-flight requests during a kamal-proxy cutover, and document the chosen Kamal `drain_timeout`/stop-grace setting in `config/deploy.yml` to match.
- [ ] T020 [US2] Run `quickstart.md` Scenario 2 end-to-end: deliberately break `/up`, merge, and confirm zero failed requests from a continuous poller while production stays on the old version; then fix `/up` and confirm a clean zero-downtime cutover.

**Checkpoint**: Deploys (success or failure) never cause downtime or ship an unhealthy version, without operator intervention.

---

## Phase 5: User Story 3 - Resilient deploys with failure notifications (Priority: P2)

**Goal**: A deploy that can't reach the production server retries with backoff instead of failing immediately, and the operator is notified via chat once retries are exhausted or a deploy/health-check fails.

**Independent Test**: Make the production server unreachable during a deploy attempt and confirm the pipeline retries with increasing delay and sends a chat notification once it keeps failing.

### Implementation for User Story 3

- [ ] T021 [US3] Wrap the `kamal deploy` step in `.github/workflows/deploy.yml` (from T010) with a bounded exponential-backoff retry for server-unreachable (SSH connection) failures.
- [ ] T022 [US3] Add a final always-run step in `.github/workflows/deploy.yml` that posts a Slack/Discord incoming-webhook Notification (using the webhook URL secret from T005), firing when retries are exhausted or the deploy/health-check failed, per `data-model.md`'s Notification entity.
- [ ] T023 [US3] Ensure the notification payload built in T022 includes `event_type` (`retries_exhausted` | `rollback_occurred`), PR number/commit SHA, and a short human-readable reason, per `contracts/ci-deploy-trigger.md`.
- [ ] T024 [US3] Run `quickstart.md` Scenario 3 end-to-end: block SSH to the server, merge, confirm backoff attempts in the workflow log and a webhook notification once exhausted; then restore access and confirm a clean success with no stale alert.

**Checkpoint**: Deploy-pipeline infrastructure failures are retried safely and surfaced to the operator without anyone needing to watch every merge.

---

## Phase 6: User Story 4 - Database migrations apply themselves on merge (Priority: P2)

**Goal**: New migrations merged into `main` are applied to production automatically before the new version goes live; a failing migration aborts the deploy and notifies the operator.

**Independent Test**: Merge a PR with a new migration and confirm the production schema reflects it after deploy, with no manual `migrate` command run.

### Implementation for User Story 4

- [ ] T025 [US4] Create `.kamal/hooks/pre-deploy` running `migrate -path migrations -database "$DATABASE_URL" up` against the production database (reusing the existing `golang-migrate` CLI/Makefile convention) before Kamal boots the new app container.
- [ ] T026 [US4] Confirm (via `quickstart.md` Scenario 4, T028) that a non-zero exit from the hook in T025 aborts the Kamal deploy entirely — the new, possibly schema-incompatible application version is never started — per FR-006.
- [ ] T027 [US4] Extend the notification step from T022 so a migration-hook failure fires a Notification with `event_type: migration_failed`, depends on T022, T025.
- [ ] T028 [US4] Run `quickstart.md` Scenario 4 end-to-end: a valid new migration auto-applies on merge; an intentionally invalid migration aborts the deploy, leaves the app un-started, and triggers a notification.

**Checkpoint**: Migrations never need to be run by hand again, and a bad migration can't take down the running app.

---

## Phase 7: User Story 5 - Reproducible server bootstrap (Priority: P2)

**Goal**: A brand-new server can be brought to full operational parity (packages, pinned Postgres version + persistent volume, backup cron) using one documented, repeatable process.

**Independent Test**: Provision a brand-new, empty server, run only the documented bootstrap process, and confirm it matches the current production server's database engine/version, persistent volume, and backup schedule — with no reliance on memory of prior manual setup.

### Implementation for User Story 5

- [ ] T029 [P] [US5] Create `ops/bootstrap.sh`: an idempotent script that runs `kamal setup` (the same operation performed manually in T008) against a target host and installs the backup cron/timer unit from `ops/backup/` (created in Phase 8, T033/T034) — turning the one-time manual bootstrap from Phase 2 into a repeatable runbook. **Also wire in `ops/postgres/provision-app-role.sh`** (added during Phase 2 as a reactive fix — the app was connecting to Postgres as the `postgres` superuser; that script creates a restricted `parkspot_app` role instead) so a fresh server gets the restricted role automatically, not just when someone remembers to run it by hand.
- [ ] T030 [US5] In `config/deploy.yml` (from T006), explicitly pin the Postgres accessory's image tag (e.g. `postgres:16.4`) and declare its named persistent Docker volume, depends on T006.
- [ ] T031 [US5] Verify that changing the pinned tag from T030 and re-running the Postgres accessory boot step (`kamal accessory reboot postgres` or equivalent) upgrades a test server to the new version without touching the persistent volume's data.
- [ ] T032 [US5] Run `quickstart.md` Scenario 5 end-to-end on a brand-new test server: `ops/bootstrap.sh` alone produces Docker, the pinned Postgres accessory with its volume, the backup cron entry, and an app container passing `/up`.

**Checkpoint**: Losing the production server is no longer a "what did I set up again?" risk — `ops/bootstrap.sh` + `config/deploy.yml` fully reproduce it.

---

## Phase 8: User Story 6 - One-command backup restore (Priority: P3)

**Goal**: The production database backs up on a recurring, reproducible schedule, and can be restored from any backup file in at most two commands.

**Independent Test**: Take an existing backup file, run the documented restore procedure against a fresh database, and confirm the restored data matches the backup.

### Implementation for User Story 6

- [ ] T033 [P] [US6] Create `ops/backup/backup.sh`: runs `pg_dump --format=custom` against the production database and uploads the resulting file to the configured object storage bucket (via `rclone` or the provider's CLI), per `data-model.md`'s Backup entity.
- [ ] T034 [P] [US6] Create `ops/backup/parkspot-backup.timer` (systemd timer, with its paired `.service` unit) or an equivalent crontab entry defining the recurring schedule that invokes `ops/backup/backup.sh`, referenced by `ops/bootstrap.sh` (T029), depends on T029, T033.
- [ ] T035 [US6] Verify the two-command `pg_restore` procedure documented in `quickstart.md` (`pg_restore --clean --if-exists -d "$DATABASE_URL" <file>`, after copying the backup file to the target) against a scratch database that already has unrelated data in it, confirming the end state is exactly the backup's data.
- [ ] T036 [US6] Run `quickstart.md` Scenario 6 end-to-end: let the scheduled backup from T034 run once, confirm the file lands in object storage, then restore it per T035.

**Checkpoint**: All six user stories are independently functional; disaster recovery is a two-command operation.

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Wrap-up work that spans multiple stories and closes out the original "it's a mess" complaint at the documentation level.

- [ ] T037 [P] Add a "Deploying" section to `README.md` pointing at `config/deploy.yml`, `.github/workflows/deploy.yml`, and `ops/bootstrap.sh` as the three files that define the entire production setup.
- [ ] T038 Run the full `quickstart.md` validation (Scenarios 1–6, in order) end-to-end against the real production (or a production-like staging) server before considering this feature done.
- [ ] T039 Identify and remove any now-superseded manual deploy notes/scripts (the old ad hoc `scp` process this feature replaces), if any exist in the repo or its docs.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately.
- **Foundational (Phase 2)**: Depends on Setup (needs `Dockerfile` from T003) — BLOCKS all user stories; nothing can be deployed or validated before T008 completes.
- **User Stories (Phase 3–8)**: All depend on Foundational completion.
  - US1 and US2 are both P1 and tightly coupled in practice (Kamal's deploy mechanism is inherently health-check-gated) — implement US1 first to get a working deploy using Kamal's default healthcheck behavior, then US2 to add the dedicated `/up` endpoint and explicitly validate the zero-downtime/rollback guarantee.
  - US3 and US4 (P2) both extend the workflow created in US1 and can proceed in either order once US1 is done; US4 depends on US3's notification step (T022) only for the `migration_failed` notification (T027) — everything else in US4 is independent.
  - US5 (P2) hardens the one-time bootstrap from Foundational (T008) into a reusable script; it can start as soon as Foundational is done, in parallel with US2–US4.
  - US6 (P3) depends on US5's bootstrap script (T029) existing before its cron/timer unit (T034) can be wired in.
- **Polish (Phase 9)**: Depends on all desired user stories being complete.

### Within Each User Story

- Config/infra tasks before their corresponding validation task.
- Each story's final task is always "run its quickstart.md scenario end-to-end" — the story isn't done until that passes.

### Parallel Opportunities

- All Setup tasks (T001–T004) can run in parallel.
- T005, T006, T007 (Foundational) touch different files/settings and can run in parallel; T008 and T009 must wait for them.
- Once Foundational is done, US5 (T029–T032) and US6 (T033–T036, minus T034's dependency on T029) can proceed in parallel with US2/US3/US4 work, since they touch `ops/` rather than `.github/workflows/deploy.yml` or `api/`.
- T015 and T017 (US2) touch different files (`api/health_handler.go` vs. `api/health_handler_test.go`) but T017 needs T015's function signature to compile against, so treat as sequential in solo work even though they're not literally the same file.

---

## Parallel Example: Foundational Phase

```bash
Task: "Configure CI/host secrets (SSH key, bws token, webhook URL) as GitHub Actions secrets"
Task: "Create config/deploy.yml with app service, production server target, and Postgres accessory"
Task: "Create .kamal/secrets resolving app env vars from GitHub secrets / bws"
```

## Parallel Example: User Story 5 & User Story 6 (once Foundational is done)

```bash
Task: "Create ops/bootstrap.sh wrapping kamal setup + backup cron install"
Task: "Create ops/backup/backup.sh running pg_dump + object storage upload"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (CRITICAL — blocks everything, includes the one real `kamal setup` run).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: run `quickstart.md` Scenario 1. Merges now auto-deploy — the core "it's a mess" problem is solved.

### Incremental Delivery

1. Setup + Foundational → a manual `kamal deploy` works.
2. Add US1 → merges auto-deploy (MVP!).
3. Add US2 → deploys are zero-downtime and self-healing on failure.
4. Add US3 → infrastructure flakiness is retried and surfaced, not silently stuck.
5. Add US4 → migrations stop being a manual step.
6. Add US5 → losing the server stops being a risk.
7. Add US6 → disaster recovery becomes a two-command operation.
8. Phase 9 → close the loop on documentation/cleanup.

---

## Notes

- [P] tasks touch different files/settings and have no completed-task dependency between them.
- [Story] labels map every Phase 3–8 task to its spec.md user story for traceability.
- Every user story phase ends in a `quickstart.md` scenario run — that is this feature's test suite.
- Commit after each task or logical group, consistent with how `specs/003-automated-deployment-pipeline`'s earlier artifacts were committed.
