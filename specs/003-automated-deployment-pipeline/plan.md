# Implementation Plan: Automated CI/CD Deployment Pipeline

**Branch**: `003-automated-deployment-pipeline` | **Date**: 2026-10-15 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/003-automated-deployment-pipeline/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Replace the current manual `scp`-based release process with a GitHub Actions workflow that triggers only on merged pull requests into `main`, builds and pushes a Docker image, and deploys it with **Kamal** (chosen over Dokku and hand-rolled scripting — see `research.md`) to the existing single production VPS that also hosts Postgres. Kamal's built-in `kamal-proxy` gives zero-downtime, health-check-gated traffic cutover (backed by a new `GET /up` endpoint) with automatic "stay on the last good version" behavior when a new release fails its health check; a Kamal pre-deploy hook runs the project's existing `golang-migrate` migrations against production before cutover; Postgres runs as a pinned-version Kamal accessory on a persistent Docker volume; a repo-tracked backup script + cron/timer unit (installed by the same bootstrap step that runs `kamal setup`) performs scheduled `pg_dump` uploads to object storage with a two-command `pg_restore` recovery path; secrets are resolved at deploy time from GitHub Actions encrypted secrets and the Bitwarden Secrets Manager CLI and injected purely as environment variables, never stored in the repo; and the workflow retries with backoff on server-unreachable failures and posts a Slack/Discord webhook notification when retries are exhausted or a deploy/migration fails.

## Technical Context

**Language/Version**: Go 1.25 (existing app, unchanged); Bash (bootstrap/backup scripts); YAML (Kamal `config/deploy.yml`, GitHub Actions workflows); Ruby gem runtime for the `kamal` CLI itself (invoked from CI, not part of the app).

**Primary Dependencies**: Kamal v2 (deploy orchestrator + bundled `kamal-proxy`), Docker Engine (build/runtime on both the CI image build step and the production host), `golang-migrate` CLI (already a project dependency/Makefile target — reused, not replaced), GitHub Actions (CI/CD platform, already hosts this repo), GitHub Container Registry (`ghcr.io`, image storage), Bitwarden Secrets Manager CLI (`bws`, app-secret resolution).

**Storage**: PostgreSQL, pinned minor version (e.g. `postgres:16.4`), run as a Kamal accessory container on the same host as the app, data on a named/persistent Docker volume. Backups: `pg_dump --format=custom` files uploaded on a schedule to S3-compatible object storage.

**Testing**: No change to the existing Go test suite (`go test`, `testcontainers-go`) — it continues to run as a required branch-protection check before merge. The deploy pipeline itself does not run tests; it is validated via the `GET /up` health check and the manual scenarios in `quickstart.md`.

**Target Platform**: Single Linux VPS reachable over SSH (current production host), with Docker installed by `kamal setup` as part of bootstrap. No Kubernetes/PaaS/multi-node orchestration.

**Project Type**: Existing web service (Go/chi backend) + new operational/CI-CD surface (workflow YAML, Kamal config, bootstrap/backup scripts) committed to the same repo — not a new application component or additional deployable service.

**Performance Goals**: No new performance target for the application itself; the pipeline's own goal is 0 seconds of full-outage downtime per deploy/rollback (SC-003) and a health check that resolves within a few seconds of container start.

**Constraints**: No secrets of any kind (plaintext or encrypted) committed to the repository; deploys and their migrations are serialized (never two in flight at once); the deploy pipeline must not re-run the test suite merged code already passed; a bounded retry/backoff budget before giving up and notifying the operator, rather than retrying forever.

**Scale/Scope**: Single production server, single Postgres instance, single application, solo operator — explicitly not designed for multi-region, multi-tenant, or multi-node scale.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

`.specify/memory/constitution.md` is still the unfilled placeholder template (no project-specific principles have been ratified yet), so there are no concrete gates to check this plan against. No violations to track; `Complexity Tracking` below is left empty accordingly. Re-checked after Phase 1 design: still no constitution in force, no new violations introduced.

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md        # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
api/                       # existing — adds one new unauthenticated handler
└── health_handler.go      # GET /up (contracts/health-check-endpoint.md)

config/                    # existing — no changes; already env-var driven

migrations/                # existing — unchanged format, now auto-applied by CI

.kamal/
├── secrets                # resolves app secrets (GitHub secrets + bws) into env vars at deploy time
└── hooks/
    └── pre-deploy          # runs `migrate ... up` against production before cutover

config/deploy.yml           # Kamal config: app image, servers, Postgres accessory (pinned
                             # version + persistent volume), proxy/healthcheck settings

.github/
└── workflows/
    └── deploy.yml           # triggers on merged PR into main; build+push image, kamal deploy
                              # with backoff retry, chat-webhook notification on failure

ops/                         # new — reproducible host bootstrap & backup, checked into the repo
├── bootstrap.sh             # idempotent: installs Docker (via `kamal setup`), installs backup
│                             # cron/timer unit; the whole "new server" runbook
└── backup/
    ├── backup.sh            # pg_dump --format=custom | upload to object storage
    └── parkspot-backup.timer / .cron    # schedule definition, installed by bootstrap.sh
```

**Structure Decision**: This feature adds no new application module or service — it is almost entirely CI/CD and host-provisioning configuration layered onto the existing single Go backend. The only application code change is one new handler (`api/health_handler.go`) wired into the existing `chi` router in `api/server.go`. Everything else (`config/deploy.yml`, `.kamal/`, `.github/workflows/deploy.yml`, `ops/`) is new top-level pipeline/ops configuration, kept out of `config/`, `api/`, `app/`, and `storage/` (the existing application layering) so the deployment concern stays clearly separated from application code.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
