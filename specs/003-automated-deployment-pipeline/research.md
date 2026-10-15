# Research: Automated CI/CD Deployment Pipeline

## 1. Deployment orchestrator: Kamal vs. Dokku vs. custom scripting

**Decision**: Use **Kamal** (37signals' Docker/SSH deploy tool, v2) as the deployment orchestrator, driven from a GitHub Actions workflow on PR-merge.

**Rationale**:
- The target environment is exactly what Kamal is built for: one (or a few) plain Linux VPS boxes reachable over SSH, no Kubernetes, no PaaS account. `kamal setup` bootstraps a brand-new server itself — installs Docker, creates the Docker network, logs in to the registry, and boots the app + accessories — directly satisfying the "new server shouldn't require remembering anything" requirement (FR-014) because the entire server state is declared in `config/deploy.yml` + `.kamal/secrets`, both checked into the repo.
- Zero-downtime and health-gated cutover are native, not bolted on: Kamal boots the new container, waits for `kamal-proxy` to see it pass its configured health check, and only then atomically switches traffic; the old container stays up and serving until that happens. If the new container never becomes healthy, traffic never moves — production keeps running the last good version with no operator action (FR-007, FR-008, FR-009). This is a better fit for "automatic rollback" than Dokku or a hand-rolled script, both of which need extra tooling to get the same guarantee.
- Postgres fits Kamal's "accessory" concept: declared once in `deploy.yml` with a pinned image tag (e.g. `postgres:16.4`) and a named Docker volume for persistent storage. Re-running `kamal accessory boot postgres` on a new box reproduces the exact same database engine/version/volume setup (FR-014, FR-016, FR-019) with no manual `apt install postgresql` step to forget.
- Kamal has first-class pre-deploy hooks (`.kamal/hooks/pre-deploy`), which is the natural place to run `golang-migrate` against the production `DATABASE_URL` before the new release is allowed to boot (FR-005, FR-006). A failing hook aborts the deploy before any traffic cutover is attempted.
- Secrets are resolved by `.kamal/secrets` at deploy time from whatever the CI runner's environment provides — GitHub Actions encrypted secrets, or a shell call out to the Bitwarden Secrets Manager CLI (`bws`) — and are injected into the container as environment variables, matching FR-012/FR-013 and the existing `config.Config` pattern (all app config already comes from env vars today, including `DATABASE_URL`, `TOKEN_ACCESSSECRET`, `OAUTH_*`; nothing in `config/` needs to change).
- Kamal is a single Ruby gem invoked from CI (`gem install kamal` + `kamal deploy`); it needs no agent installed on the server ahead of time beyond SSH + Docker, so there's no bespoke "remember what I installed" surface beyond what `deploy.yml` already declares.

**Alternatives considered**:
- **Dokku** — a real, actively maintained option, and like Kamal it turns a plain VPS into something `git push`-deployable with its own nginx-based proxy and zero-downtime support via a `CHECKS` file. Rejected as the primary choice because: (a) Dokku's reproducibility story is more imperative than declarative — most of its configuration (`dokku config:set`, `dokku postgres:create`, plugin installs) is applied by running CLI commands against a live server rather than captured in one file in the repo, so "what did I set up on the old server" risk only moves from "forgot the cron job" to "forgot which `dokku` commands I ran," unless we additionally hand-write and maintain an idempotent wrapper script — which is exactly the custom-scripting option below, just with Dokku underneath it; (b) automatic rollback-on-failed-health-check is not built in the way Kamal's is — Dokku will refuse to cut over traffic to a container that fails its `CHECKS`, which covers FR-008, but recovering a "I force-deployed something bad anyway" situation needs the community `dokku-rollback` plugin, an extra moving part; (c) Dokku's Postgres is normally managed via the separate `dokku-postgres` plugin, another thing to install and keep pinned per-server. Dokku remains a reasonable fallback if Kamal's Docker-registry-centric model ever becomes a blocker (e.g., wanting `git push`-style deploys instead of CI-built images).
- **Custom scripting** (GitHub Actions → SSH → `docker compose up -d` or raw systemd units, hand-rolled blue/green swap behind Caddy/Nginx, hand-rolled rollback logic, hand-rolled bootstrap shell/Ansible script) — rejected as the primary path because every one of the hard requirements this feature exists to deliver (zero-downtime cutover, health-gated rollback, backoff+notify, reproducible bootstrap) would need to be designed and debugged from scratch and kept correct forever by a solo maintainer, which is the “mess” this feature is meant to replace. It remains the fallback if Kamal turns out not to fit once implementation starts.
- **Nomad / Docker Swarm / k3s** — rejected as disproportionate: all solve problems (multi-node scheduling, service discovery across many hosts) this single-server, single-app, single-maintainer project doesn't have, and each adds its own "new server setup" surface this feature is trying to shrink, not grow.

## 2. CI/CD platform and trigger

**Decision**: **GitHub Actions**, with a workflow that triggers only `on: pull_request: types: [closed]` filtered to `github.event.pull_request.merged == true` and base branch `main` (equivalently, `push` to `main` gated by checking the triggering commit is a merge commit produced by a PR, via the GitHub API/`github.event.pull_request`). Branch protection on `main` (required status checks, no direct pushes) is configured in the repository settings, not in workflow code.

**Rationale**: The repository already lives on GitHub (`origin` remote is `github.com:dannyvelas/parkspot-backend`), so GitHub Actions needs no new hosting account, has first-class `pull_request`/`push` event metadata to distinguish "merged PR" from "direct push," and has built-in encrypted repository/environment secrets for the one credential Kamal truly needs out of CI (the SSH deploy key / `KAMAL_REGISTRY_PASSWORD` / Bitwarden access token) without checking anything into the repo.

**Alternatives considered**: GitLab CI, CircleCI, Drone — all viable but would require moving the repo or paying for/hosting a runner for no added benefit over GitHub Actions given the repo is already on GitHub.

## 3. Container registry

**Decision**: **GitHub Container Registry (ghcr.io)**, authenticated via the workflow's built-in `GITHUB_TOKEN`.

**Rationale**: Zero additional account/billing setup, integrates with the existing GitHub repo and Actions auth, and is a first-class target for Kamal (`kamal` builds the image and pushes/pulls it as part of `kamal deploy`).

**Alternatives considered**: Docker Hub (extra account, rate limits on the free tier), self-hosted registry (one more service to keep reproducible on the same box that's supposed to be getting simpler).

## 4. Secrets delivery

**Decision**: GitHub Actions **encrypted repository/environment secrets** hold the one bootstrapping credential CI needs (SSH private key to reach the production host, plus registry auth). Application-level secrets (DB credentials, `TOKEN_ACCESSSECRET`/`TOKEN_REFRESHSECRET`, `OAUTH_*`) are resolved by `.kamal/secrets` at deploy time by shelling out to the **Bitwarden Secrets Manager CLI (`bws`)**, which the maintainer already uses, and are passed into the container purely as environment variables — matching the existing `config/read_env.go` pattern with no application code changes.

**Rationale**: Satisfies FR-012/FR-013 exactly as scoped: nothing encrypted or plaintext ever lives in the repo; both GitHub's native secret store and Bitwarden are legitimate "secrets manager" options the user already named as acceptable, and `.kamal/secrets` supports both natively (plain env-var interpolation for the GitHub-held bootstrap credential, and arbitrary shell commands for pulling the rest from `bws`).

**Alternatives considered**: SOPS/age-encrypted files in the repo — explicitly ruled out by the user. A dedicated external secrets operator/sidecar (e.g., Vault, Infisical) — rejected as unnecessary operational overhead for a single-server, single-app deployment when Bitwarden already covers the need.

## 5. Zero-downtime HTTP cutover and health check

**Decision**: Rely on **kamal-proxy** (bundled with Kamal) in front of the Go service, pointed at a new, dedicated unauthenticated `GET /up` endpoint added to the API that returns `200 OK` once the process can serve traffic (e.g., after confirming it can reach Postgres). `main.go` already performs graceful shutdown on `SIGINT`/`SIGTERM` (`server.ShutdownGracefully(30 * time.Second)`), which is exactly what's needed on the *old* container's side of a Kamal cutover (it keeps draining in-flight requests after traffic stops being routed to it, rather than being killed immediately).

**Rationale**: kamal-proxy only routes traffic to a container once it's confirmed healthy and only stops sending traffic to the old one once the new one is confirmed, which is what makes FR-009 ("no window of complete unavailability") achievable without hand-written proxy logic. The existing public, unauthenticated `/` route ("hello world") is not reused because it doesn't check any real dependency (e.g., DB connectivity) and mixing a load-bearing health check into an unrelated legacy route would make intent unclear.

**Alternatives considered**: Reusing the existing authenticated `/hello` route — rejected, it requires a valid access token, which a health check should not depend on. Caddy/Nginx hand-rolled upstream health checks — rejected per Decision 1 (Kamal already includes this).

## 6. Database migrations in the pipeline

**Decision**: Keep **`golang-migrate`** (already a project dependency and the tool behind the existing `Makefile` `migrate_*` targets and the `migrations/*.up.sql`/`*.down.sql` files) as the migration tool. Run `migrate -path migrations -database "$DATABASE_URL" up` from a Kamal `pre-deploy` hook, executed against production before the new application container is allowed to boot.

**Rationale**: No new migration tooling or file format is introduced — this is purely automating a command the maintainer already runs by hand today. Running it as a `pre-deploy` hook means a failing migration aborts the deploy before the new (possibly schema-incompatible) code is ever started, satisfying FR-006.

**Alternatives considered**: Running migrations from inside the application's own startup path — rejected because it would run on every container start (including rollbacks/restarts), making it harder to reason about exactly-once application of a migration, and because a migration failure would then surface as an application crash-loop instead of a clean, pre-flight pipeline failure.

## 7. Backoff, retries, and chat notifications

**Decision**: Wrap the `kamal deploy` step in the GitHub Actions workflow with a bounded exponential-backoff retry (e.g., a small number of attempts with increasing delay) for the specific failure mode of "can't reach the server at all" (SSH connection failure). Once retries are exhausted, and separately whenever Kamal reports that a deploy's health check failed (i.e., the new version never went live), post a message to a **Discord incoming webhook** (or Slack incoming webhook — both are simple HTTPS POSTs with no app to install, interchangeable at the workflow-config level) from a final always-run step in the workflow.

**Rationale**: Satisfies FR-010/FR-011 and the backoff-exhaustion edge case in the spec. Because Kamal never cuts traffic over to an unhealthy container, retries are inherently safe to repeat — they cannot make production worse, only delay recovery of the *deploy pipeline* itself, not of production traffic.

**Alternatives considered**: A dedicated uptime/alerting SaaS (e.g., PagerDuty) — rejected as disproportionate for a solo-maintainer project when a webhook post is sufficient and the user already lives in chat apps.

## 8. Reproducible server bootstrap, pinned Postgres, and persistent volume

**Decision**: Postgres runs as a **Kamal accessory** with a pinned image tag (`postgres:<MAJOR.MINOR>`, e.g. `postgres:16.4`) and a named Docker volume for `/var/lib/postgresql/data`, declared in `config/deploy.yml`. A fresh server is brought to parity by running `kamal setup` (installs Docker, creates the app + accessory containers) once, which is the entire "new server" runbook.

**Rationale**: Directly satisfies FR-014/FR-016/FR-019 and User Story 5's independent test ("provision a brand-new server, run only the documented bootstrap process"): the declarative `deploy.yml` plus `kamal setup` is the single source of truth, so there is nothing left to "remember."

**Alternatives considered**: A managed Postgres service — explicitly out of scope per the spec (DB stays colocated on the same host as the app). A separate Ansible/Terraform provisioning layer — unnecessary extra moving part when Kamal's own setup step already covers the whole surface this feature needs.

## 9. Reproducible backup cron

**Decision**: A backup script (`pg_dump` in custom format, piped to an object-storage upload via `rclone` or the target provider's CLI) plus its cron schedule are both defined as files checked into the repo (the script itself, and a crontab/systemd-timer unit), installed onto the host as part of the same bootstrap step that runs `kamal setup` — not configured by hand afterward.

**Rationale**: Directly satisfies FR-015/FR-017 and the spec's explicit "the cron ... is reproducible and not something I'll forget to set up." Treating the cron definition as a file in version control, applied by the same bootstrap step, means a new server gets the backup job "for free" instead of as a separately-remembered manual step.

**Alternatives considered**: Driving backups from a scheduled GitHub Actions workflow (`on: schedule`) that SSHes in to trigger `pg_dump` — rejected because it adds a second, less reliable trigger path (depends on GitHub Actions' cron being timely, and still needs the same SSH access) for no benefit over a plain host-local cron/systemd-timer that's already right next to the database.

## 10. Backup restore procedure

**Decision**: Document (in `quickstart.md`) a two-command restore: (1) copy/point the chosen backup file at the server, (2) run `pg_restore` (matching the `pg_dump` custom format chosen in Decision 9) against the target database, optionally inside the running Postgres accessory container via `kamal accessory exec postgres -- ...` or a plain `docker exec`.

**Rationale**: Satisfies FR-018/SC-008 ("no more than two commands") using the standard tool that pairs with `pg_dump`, with no new bespoke restore tooling to build or maintain.

**Alternatives considered**: A custom restore script wrapping `pg_restore` — left as an optional future convenience, not required to meet the "two commands" bar since plain `pg_restore` already clears it.
