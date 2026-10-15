# Quickstart: Validating the Automated Deployment Pipeline

This is a validation guide, not implementation code. It assumes the design in `plan.md`/`research.md`/`data-model.md` has been implemented: a `config/deploy.yml` (Kamal), a `GET /up` endpoint (`contracts/health-check-endpoint.md`), a GitHub Actions workflow (`contracts/ci-deploy-trigger.md`), a backup script + cron unit, and branch protection on `main`.

## Prerequisites

- A test Linux server (or a fresh VM) reachable over SSH, with a DNS name/IP you control — this stands in for "a brand-new production server."
- `kamal` installed locally or available in the CI runner image.
- The repo's GitHub Actions secrets configured: SSH deploy key, `ghcr.io` auth (via `GITHUB_TOKEN`), Bitwarden Secrets Manager access token (or the equivalent GitHub Actions secrets standing in for app secrets), and a Slack or Discord webhook URL.
- Branch protection enabled on `main` requiring the existing test-suite status check.

## Scenario 1 — Merge-to-deploy (User Story 1)

1. Open a trivial PR (e.g., a comment or log-line change) against `main` and get its required check green.
2. Merge the PR.
3. **Expect**: the GitHub Actions deploy workflow starts automatically with no manual trigger; it finishes with the new commit live.
4. Push a commit directly to `main` via a local branch merge bypass (if branch protection allows it in a test repo) or simulate via a non-PR `push` event.
5. **Expect**: no deployment workflow run is triggered by step 4.

## Scenario 2 — Health-gated zero-downtime deploy + automatic rollback (User Story 2)

1. While a small script continuously polls a real (authenticated or public) endpoint of the running service every ~200ms, merge a PR whose new container is deliberately broken (e.g., temporarily make `/up` return `503`).
2. **Expect**: the polling script records zero failed requests throughout — the old container keeps serving until kamal-proxy gives up on the new one; the workflow run ends in a failure state; production is confirmed still running the previous commit.
3. Revert the deliberate breakage, merge again.
4. **Expect**: the new version passes `/up`, traffic cuts over with zero failed requests in the poller, and the workflow run ends successfully.
5. Check the configured chat channel: a Notification was posted for the failed attempt in step 2.

## Scenario 3 — Unreachable server, backoff, and notification (User Story 3)

1. Temporarily block SSH access to the test server (e.g., firewall rule) and merge a PR.
2. **Expect**: the workflow log shows multiple deploy attempts with increasing delay between them, then a chat Notification once the retry budget is exhausted.
3. Restore SSH access and re-run the workflow (or merge again).
4. **Expect**: the deploy succeeds and no stale failure notification is left unresolved/misleading.

## Scenario 4 — Migrations apply automatically (User Story 4)

1. Add a trivial new migration file pair (`NNNNNN_quickstart_check.up.sql` / `.down.sql`) to a PR and merge it.
2. **Expect**: after the deploy workflow completes, the migration's effect is visible in the production schema (e.g., via `migrate version` against the production `DATABASE_URL`) with no manual `migrate` command run by the operator.
3. Repeat with a migration file containing intentionally invalid SQL.
4. **Expect**: the deploy workflow fails at the migration step, the application container is never started/swapped in, and a Notification is sent.

## Scenario 5 — Fresh server bootstrap (User Story 5)

1. Provision a brand-new, empty test server.
2. Run only the documented bootstrap command(s) (`kamal setup` against this new host, pointed at the same `config/deploy.yml`) — no other manual configuration.
3. **Expect**: Docker is installed, the pinned Postgres accessory is running with its persistent volume, the backup cron/timer is present (`crontab -l` or `systemctl list-timers`), and the app container is running and passing `/up`.
4. Change the pinned Postgres image tag in `config/deploy.yml` and re-run the accessory boot step.
5. **Expect**: the new server now runs the newly pinned version, confirmed via `docker exec <postgres accessory> postgres --version`.

## Scenario 6 — Backup and restore (User Story 6)

1. Let the backup cron run once (or trigger it manually) and confirm a new dump file lands in the configured object storage bucket.
2. On a scratch database, run the documented restore procedure (at most two commands) pointed at that dump file.
3. **Expect**: the scratch database's contents match what was in production at dump time.
4. Repeat the restore against a scratch database that already has unrelated data in it.
5. **Expect**: the end state is exactly the backup's data, not a merge of old and restored rows.
