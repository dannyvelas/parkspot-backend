# Contract: CI → Deploy Trigger

This documents the external trigger/notification surfaces the pipeline itself exposes (as opposed to the application's own HTTP API).

## Trigger: merged pull request into `main`

- **Fires on**: a GitHub `pull_request` event with `types: [closed]` where `github.event.pull_request.merged == true` and `github.event.pull_request.base.ref == 'main'` (equivalently, a `push` to `main` whose head commit is a merge produced by a PR — either formulation is acceptable at implementation time, but the workflow MUST NOT fire a deployment for a `push` to `main` that is not the result of a merged PR).
- **Does not fire on**: direct pushes to `main` that bypass a pull request; pushes to any other branch; PR `closed` events where `merged == false` (i.e., a closed-without-merging PR).
- **Input available to the job**: merge commit SHA, PR number, PR base/head refs — used to tag the built image (`image_ref`) and to label any resulting Notification.
- **No re-validation**: the triggered job MUST NOT re-run the test suite; branch protection on `main` is the sole gate for code quality, enforced before merge, not by this workflow.

## Output: chat Notification (Slack/Discord incoming webhook)

- **Fires on**: deploy retries exhausted while the server was unreachable; a deploy whose health check failed (traffic never cut over); a migration that failed to apply.
- **Payload**: a plain text/JSON message (per the target webhook's own format) containing at minimum: event type, PR number/commit SHA, and a short human-readable reason.
- **Delivery**: a single HTTPS POST per event to a webhook URL held as a CI secret; no retry of the notification itself is required (if the notification POST fails, that is a secondary, lower-severity problem than the deploy failure it was reporting).
