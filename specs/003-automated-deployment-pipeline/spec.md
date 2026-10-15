# Feature Specification: Automated CI/CD Deployment Pipeline

**Feature Branch**: `003-automated-deployment-pipeline`

**Created**: 2026-10-15

**Status**: Draft

**Input**: User description: "this is the repo of the backend of the parkspot web application. i want my merges to main to automatically become deployments to production. i want something that is easily reproducible. if i get a new server, i dont want to have to remember \"wait what was the cron job i had running in my old server again?\" or \"wait what packages did i have in my old server again?\" the more plug-and-play the better. i have 0 ci/cd set up right now. i'm just manually copying files with scp to my server. it's a mess. some things to know: my production database isnt a managed service, its also not running on a dedicated host - it runs on the same server as my parkspot go service. it would be nice if: only merges from PRs to main get deployed, not just commits to main; whenever migrations get merged to main, they automatically get applied to the production database; the cron that does my database backup to object storage is reproducible and not something i'll just forget to set up in a new server; its easy to run postgres with a persistent volume and a pinned version; its easy to restore data with a command or two that i can point at my pg_dump file; if there's a problem deploying a new version that has just been merged (e.g. like it wont start), it automatically gets rolled back to the last working version without an operator needing to intervene; if there's a problem in the deploying (e.g. like the production server can't be reached), it should do a back-off in trying to deploy and i should get notified via slack or discord or something similar; after a deployment, there should be a healthcheck to some http endpoint or something that validates that the new deployment is working as expected; i dont want to keep encrypted secrets in the repo, i would rather my ci/cd pipeline somehow keep secrets and inject them as env variables, i would be open to a secrets manager that injects secrets into my app (i already use bitwarden); i want zero downtime deploys, a few seconds of downtime during deploys is not okay; PRs must pass CI checks before they're even mergeable (branch protection); the deploy pipeline trusts merged code already passed tests and does not re-run them (even if an admin force-merged a PR)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Merge to main automatically ships to production (Priority: P1)

As the sole maintainer of the ParkSpot backend, when my pull request is merged into `main`, I want the new version to go live in production automatically, without me manually building anything or copying files over `scp`.

**Why this priority**: This is the core problem statement — today every release is a manual, error-prone, undocumented `scp` session. Removing that manual step is the single highest-value change and is a prerequisite for every other story in this feature.

**Independent Test**: Merge a trivial, approved PR into `main` and confirm the running production service reflects the new code without any manual deployment action being taken.

**Acceptance Scenarios**:

1. **Given** a pull request has passed all required CI checks and is merged into `main`, **When** the merge completes, **Then** a deployment to production starts automatically with no manual trigger.
2. **Given** a commit is pushed directly to `main` (not via a merged pull request), **When** that push happens, **Then** no production deployment is triggered by it.
3. **Given** an admin merges a pull request without its CI checks passing (e.g., force-merge), **When** the deploy pipeline runs, **Then** it deploys the merged code without re-running the test suite, treating the merge as already validated.
4. **Given** a deployment has just completed, **When** I inspect the deployment history, **Then** I can see which commit/PR is currently live and the outcome (success/failed/rolled back) of the most recent attempts.

---

### User Story 2 - Safe deploys with automatic health validation and rollback (Priority: P1)

As the maintainer, I want every automatic deployment to be validated against a live health check and, if the new version fails to start or fails that check, to be rolled back to the last known-good version without me having to notice or intervene — and I want this to happen with no visible downtime for end users.

**Why this priority**: Automatic deploys are only safe to adopt if a bad release can't take production down unattended. This is what makes Story 1 trustworthy enough to actually rely on.

**Independent Test**: Deploy a version that deliberately fails to start (or fails its health endpoint), and confirm production automatically reverts to the previously running version with no end-user-visible downtime and no operator action.

**Acceptance Scenarios**:

1. **Given** a new version has been deployed, **When** the system checks its health endpoint, **Then** the deployment is only considered successful once the new version responds as healthy.
2. **Given** a newly deployed version fails to start or fails its health check, **When** this failure is detected, **Then** production automatically reverts to running the last version that was confirmed healthy, without requiring an operator to act.
3. **Given** a deployment (success or automatic rollback) is in progress, **When** end users send requests to the production service during that window, **Then** every request is still served successfully (no dropped connections, no error responses caused by the deploy itself).
4. **Given** an automatic rollback has occurred, **When** I next check my notifications, **Then** I have been informed that a deployment failed and was rolled back.

---

### User Story 3 - Resilient deploys with failure notifications (Priority: P2)

As the maintainer, if the pipeline cannot reach the production server to deploy at all (e.g., the server is temporarily unreachable), I want it to retry with backoff instead of failing immediately, and I want to be notified through a chat app I already use if it keeps failing, so I find out about infrastructure problems without having to babysit every merge.

**Why this priority**: Distinct failure mode from Story 2 (can't reach the box at all vs. the new version is unhealthy once reached). Important for trust in the system, but less critical than the core deploy and rollback behavior.

**Independent Test**: Make the production server unreachable during a deploy attempt and confirm the pipeline retries with increasing delay between attempts and sends a chat notification if it continues to fail.

**Acceptance Scenarios**:

1. **Given** a deployment attempt cannot reach the production server, **When** the attempt fails, **Then** the pipeline automatically retries after a delay, with the delay increasing on each subsequent failure.
2. **Given** retries continue to fail past a reasonable threshold, **When** that threshold is reached, **Then** I receive a notification in a chat application (e.g., Slack or Discord) describing the failure.
3. **Given** a retried deployment attempt eventually succeeds, **When** it succeeds, **Then** I am not left with unresolved failure alerts from the earlier failed attempts.

---

### User Story 4 - Database migrations apply themselves on merge (Priority: P2)

As the maintainer, when a pull request containing new database migrations is merged into `main`, I want those migrations to be applied automatically to the production database as part of the deployment, so I never have to remember to run them by hand.

**Why this priority**: Forgetting to run a migration (or running it against the wrong database) is a common source of production incidents for solo-maintained services. This builds directly on Story 1's deploy trigger.

**Independent Test**: Merge a PR that adds a new migration and confirm the production database schema reflects that migration after the deploy completes, with no manual database command run by the operator.

**Acceptance Scenarios**:

1. **Given** a merged PR includes one or more new database migrations, **When** the deployment for that merge runs, **Then** the pending migrations are applied to the production database automatically, before or as part of bringing up the new version.
2. **Given** a merged PR includes no new migrations, **When** the deployment runs, **Then** no migration step changes the database, and the deploy proceeds normally.
3. **Given** a migration fails to apply, **When** that failure happens, **Then** the deployment does not proceed to run the new, incompatible version against the database, and I am notified of the failure.

---

### User Story 5 - Reproducible server bootstrap (Priority: P2)

As the maintainer, if I ever need to move to a new server, I want to be able to stand up an equivalent production environment — OS packages, the database engine at the same pinned version with its data volume, the backup cron job, and the deploy target — by running a documented, repeatable process, instead of relying on my memory of what I set up on the old server.

**Why this priority**: Directly addresses the "what if I get a new server" fear named in the request. It doesn't block day-to-day merges (Stories 1-4) but removes a standing risk of losing institutional knowledge about the production setup.

**Independent Test**: Provision a brand-new, empty server, run only the documented bootstrap process against it, and confirm it ends up with the same database engine/version, the same persistent data volume configuration, and the same backup cron schedule as the existing production server — without consulting any notes, chat history, or memory of prior manual setup.

**Acceptance Scenarios**:

1. **Given** a brand-new server with nothing installed, **When** the bootstrap process is run against it, **Then** all required packages and services (including the pinned database version) end up installed and configured identically to the current production server.
2. **Given** the bootstrap process has completed on a new server, **When** I check for the scheduled database backup job, **Then** it is already present and correctly configured, with no separate manual step required.
3. **Given** the database is bootstrapped on a new server, **When** it is started, **Then** its data is stored on a persistent volume that survives a restart of the database process/container.
4. **Given** I change the pinned database version in the bootstrap configuration, **When** I re-run the bootstrap process, **Then** the new server runs that version instead of requiring me to recall and manually install it.

---

### User Story 6 - One-command backup restore (Priority: P3)

As the maintainer, I want to restore the production database from one of my backup files using at most a couple of commands, so that recovering from data loss or standing up a new environment from a known data set is fast and doesn't require me to improvise a procedure under pressure.

**Why this priority**: Important for disaster recovery, but it's a less frequent operation than every-day deploys and depends on the backup mechanism from Story 5 already existing.

**Independent Test**: Take an existing backup file, run the documented restore procedure against a fresh database, and confirm the restored database matches the data captured in that backup.

**Acceptance Scenarios**:

1. **Given** a database backup file, **When** I run the documented restore procedure pointed at that file, **Then** the target database ends up containing the data from that backup, using no more than two commands.
2. **Given** a restore is run against a database that already has data in it, **When** the restore completes, **Then** the resulting state is the backup's data (not a partial merge of old and restored data), so the outcome is predictable.

---

### Edge Cases

- What happens if a second PR is merged into `main` while a deployment from an earlier merge is still in progress? Deployments MUST be processed to completion one at a time, in merge order, so two deploys never apply migrations or restart the service concurrently.
- What happens if the health check endpoint itself is unreachable or misbehaves even though the application is otherwise working correctly? This MUST be treated the same as a failed health check (triggers rollback), since the pipeline has no other way to confirm the new version is good.
- What happens if a migration partially succeeds before failing (e.g., multi-statement migration)? The deployment MUST NOT bring up the new application version on top of a database left in an unknown/partial schema state, and the operator MUST be notified.
- What happens if the automatic rollback target (the "last known-good version") is itself unavailable or also fails to start? The operator MUST be notified immediately rather than the system silently giving up or looping forever.
- What happens if the backoff/retry budget for an unreachable server is exhausted? The pipeline MUST stop retrying automatically after notifying the operator, rather than retrying forever.
- What happens to in-flight requests during the moment the new version takes over from the old version? They MUST complete successfully against whichever version was serving them, with no dropped connections.
- What happens if a secret required by the application is missing or invalid at deploy time? The deployment MUST fail before the new version is exposed to production traffic, rather than starting in a broken state.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST automatically start a production deployment whenever a pull request is merged into `main`.
- **FR-002**: The system MUST NOT treat a direct commit/push to `main` that did not come through a merged pull request as a deployment trigger.
- **FR-003**: The system MUST prevent pull requests from being merged into `main` unless they have passed required CI checks (branch protection), except where an authorized admin explicitly bypasses that protection.
- **FR-004**: The deployment pipeline MUST NOT re-run the test suite during deployment; it MUST treat code that reached `main` as already validated, including code merged via an admin bypass.
- **FR-005**: The system MUST apply any new database migrations included in a merge to the production database automatically as part of that deployment, with no manual step from the operator.
- **FR-006**: The system MUST abort the deployment (and not start the new application version) if a migration fails to apply, and MUST notify the operator of the failure.
- **FR-007**: The system MUST validate every new deployment against an HTTP health check before considering the deployment successful.
- **FR-008**: The system MUST automatically roll back to the last known-good version, without operator intervention, if the new version fails to start or fails its health check.
- **FR-009**: The system MUST perform deploys (including rollbacks) such that end-user-facing requests continue to be served successfully throughout the transition, with no window of complete unavailability.
- **FR-010**: The system MUST retry a deployment with increasing delay between attempts (backoff) when it cannot reach the production server, rather than failing immediately on the first unreachable attempt.
- **FR-011**: The system MUST notify the operator through a chat application (e.g., Slack or Discord) when deployment retries are exhausted without success, and separately when an automatic rollback occurs.
- **FR-012**: The system MUST NOT require any application secret (credentials, keys, tokens) to be stored in the source repository, encrypted or otherwise.
- **FR-013**: The system MUST inject required secrets into the running application as environment variables at deploy time, sourced from a secrets store outside the repository.
- **FR-014**: The system MUST provide a documented, repeatable process that brings a brand-new server to the same operational state as the current production server: required packages, the database engine pinned to a specific version, the database's persistent storage, and the scheduled backup job.
- **FR-015**: The scheduled database backup job MUST be defined as part of the reproducible server setup (not configured by hand on a running server), so that bootstrapping a new server automatically includes it.
- **FR-016**: The production database's data MUST be stored on a persistent volume that survives restarts of the database process/container independent of the application's own lifecycle.
- **FR-017**: The database backup job MUST upload each backup to object storage on a recurring schedule without requiring manual action once set up.
- **FR-018**: The system MUST provide a restore procedure that, given a backup file, fully restores a database to that backup's contents using no more than two operator-run commands.
- **FR-019**: The system MUST make it possible to change the pinned database version in configuration and have that version take effect the next time the database environment is (re-)provisioned.
- **FR-020**: The system MUST record, for each deployment attempt, at minimum: which commit/PR it corresponds to, its outcome (succeeded / failed / rolled back), and when it happened, in a way the operator can review after the fact.

### Key Entities

- **Deployment**: A single attempt to ship a merged commit to production. Has an associated commit/PR reference, a status (in progress, succeeded, failed, rolled back), timestamps, and the health-check outcome that determined success or failure.
- **Migration**: A versioned database schema change shipped in a PR. Has an identity/ordering relative to other migrations and an applied/not-applied state against the production database.
- **Backup**: A point-in-time export of the production database stored in object storage. Has a creation timestamp and a location/identifier that can be referenced during a restore.
- **Secret**: A piece of sensitive configuration (credential, key, token) the application needs at runtime. Lives outside the repository and is injected into the running application as an environment variable at deploy time.
- **Notification**: A message sent to the operator's chat application describing a pipeline event that needs attention (exhausted retries, automatic rollback, migration failure).
- **Server Environment**: The reproducible set of packages, pinned database version, persistent data volume, and scheduled jobs (backup) that constitute a production host, independent of any single physical/virtual machine.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of pull requests merged into `main` result in production running that code, with zero manual deployment steps performed by the operator.
- **SC-002**: Commits pushed directly to `main` outside of a merged pull request never trigger a production deployment (0 occurrences).
- **SC-003**: During deploys and automatic rollbacks, end users experience no period of complete service unavailability — 0 seconds of full downtime, measured across all deploys in a rolling 30-day window.
- **SC-004**: 100% of deployments that fail to start or fail their health check are automatically reverted to the last known-good version without any operator action, within a few minutes of the failure being detected.
- **SC-005**: 100% of merged migrations are applied to the production database automatically, with zero instances of an operator needing to run a migration command by hand.
- **SC-006**: When the production server is temporarily unreachable, the operator is notified via chat within a bounded, short time window after retries are exhausted, in 100% of such incidents.
- **SC-007**: A new server can be brought to full operational parity (packages, pinned database version, persistent volume, backup cron) using only the documented bootstrap process, with zero reliance on the operator's memory of prior manual configuration.
- **SC-008**: Restoring a database from a backup file requires at most 2 commands and completes successfully, verified by at least one practice restore.
- **SC-009**: Zero application secrets are present in the source repository (in plaintext or encrypted form) at any point after this feature is in place.

## Assumptions

- The ParkSpot backend and its production database continue to run colocated on a single server; this feature does not introduce a managed database service or split them onto separate hosts.
- "Operator" and "maintainer" in this spec refer to the single developer who currently performs deploys manually; there is no separate deployment-approval role to design for.
- A chat notification channel (Slack, Discord, or an equivalent the operator already uses) is available and is the correct place to surface pipeline alerts; no additional notification channel (email, SMS, paging) is in scope.
- The secrets store used to inject environment variables at deploy time is either the CI/CD platform's own encrypted secrets feature or an external secrets manager (e.g., Bitwarden) — either satisfies "no secrets in the repo"; the specific choice is a delivery detail, not a scoping concern for this spec.
- "Last known-good version" for rollback purposes means the most recent version that previously passed its own health check in production, not an arbitrarily older release.
- A reasonable, bounded number of retry attempts (with increasing backoff) is acceptable before the pipeline gives up and notifies the operator; the exact count/timing is a delivery detail.
- Database migrations in this project are forward-only as part of a normal deploy (i.e., this feature automates applying new migrations, not automatically reverting migrations as part of a rollback); rolling back the database schema itself remains a manual, deliberate action if ever needed.
- Backups are taken on a recurring schedule (e.g., daily) rather than continuously; the exact frequency is a delivery detail, not a scoping concern for this spec.
