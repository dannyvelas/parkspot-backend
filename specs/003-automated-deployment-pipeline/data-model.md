# Data Model: Automated CI/CD Deployment Pipeline

This feature is operational/infrastructure automation, not a new application data domain — there are no new database tables. The entities below are the concrete, durable artifacts the pipeline creates and manipulates, mapped from the spec's Key Entities to how they are represented with the tools chosen in `research.md`.

## Deployment

Represents one attempt to ship a merged commit to production.

- **Identity**: the merged commit SHA (and associated PR number) plus the CI run ID that drove it.
- **Fields**:
  - `commit_sha` / `pr_number` — what's being deployed.
  - `status` — `in_progress` | `succeeded` | `failed` | `rolled_back` (rolled_back here means "the new version never went live; production stayed on the previous release").
  - `started_at` / `finished_at`.
  - `health_check_result` — pass/fail outcome from the `/up` endpoint that gated traffic cutover.
  - `image_ref` — the `ghcr.io` image tag built and deployed for this commit.
- **Representation**: the GitHub Actions workflow run (status, logs, timestamps) is the record of truth; `image_ref` is the Docker tag Kamal built/pushed/deployed for that run. No separate database table is introduced for this.
- **Relationships**: produced by exactly one merged PR; may apply zero or more Migrations; on failure, may trigger one or more Notifications.

## Migration

A versioned database schema change shipped in a PR.

- **Identity**: the sequence number encoded in the existing `migrations/NNNNNN_name.up.sql` / `.down.sql` filenames (unchanged — this feature automates running them, it does not change their format).
- **Fields**: `version` (from filename), `applied_at`, `applied` (tracked by `golang-migrate`'s own `schema_migrations` table in Postgres — not a new entity this feature introduces).
- **State transitions**: `pending` → `applied` (via the Kamal pre-deploy hook) or `pending` → `failed` (hook aborts the Deployment; the version is left exactly as `golang-migrate` leaves a failed migration today — no new failure-handling semantics are introduced beyond what the existing CLI already does).
- **Relationships**: belongs to the Deployment whose merge introduced it; applied before that Deployment's application container is allowed to start.

## Backup

A point-in-time export of the production database.

- **Identity**: timestamped filename (e.g., `parkspot-YYYYMMDDTHHMMSS.dump`).
- **Fields**: `created_at`, `object_storage_location` (bucket/key), `size`.
- **Representation**: a `pg_dump --format=custom` file uploaded to object storage by the host's cron/systemd-timer job; no index/catalog beyond the object storage bucket's own listing is introduced.
- **Relationships**: independent of any single Deployment; consumed by the restore procedure (Decision 10 in `research.md`).

## Secret

A piece of sensitive runtime configuration.

- **Identity**: the environment variable name the application already reads (e.g., `DATABASE_URL`, `TOKEN_ACCESSSECRET`, `TOKEN_REFRESHSECRET`, `OAUTH_CLIENTID`, `OAUTH_CLIENTSECRET`, `OAUTH_ACCESSTOKEN`, `OAUTH_REFRESHTOKEN`, plus any Kamal/registry/SSH credential CI itself needs).
- **Fields**: `name`, `source` (GitHub Actions encrypted secret, or Bitwarden Secrets Manager item resolved via `.kamal/secrets`), `injected_as` (always a plain env var into the container).
- **Relationships**: resolved once per Deployment, at deploy time; never persisted to disk in the repo or on the server outside the running container's environment.

## Notification

A message sent to the operator's chat app describing a pipeline event needing attention.

- **Identity**: implicit — one message per triggering event.
- **Fields**: `event_type` (`retries_exhausted` | `rollback_occurred` | `migration_failed`), `commit_sha`/`pr_number`, `message_body`, `sent_at`.
- **Representation**: an HTTPS POST to a Slack or Discord incoming-webhook URL (itself a Secret) from the final step(s) of the GitHub Actions workflow.
- **Relationships**: triggered by a Deployment or Migration reaching a failure state.

## Server Environment

The reproducible set of packages, pinned database version, persistent data volume, and scheduled jobs that constitute a production host.

- **Identity**: the `config/deploy.yml` (Kamal config) + bootstrap/backup-cron files checked into the repo, as applied to a given host by `kamal setup`.
- **Fields**: `postgres_image_tag` (pinned version), `volume_name` (persistent data volume), `backup_schedule` (cron/timer expression), `installed_packages` (Docker + whatever Kamal's setup installs).
- **Relationships**: hosts zero or more Deployments over time; hosts exactly one Postgres accessory (with its Backups); hosts the backup cron job.
