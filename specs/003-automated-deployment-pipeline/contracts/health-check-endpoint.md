# Contract: Deployment Health-Check Endpoint

This is the one new interface the application exposes as part of this feature: the endpoint kamal-proxy polls to decide whether a newly deployed container is allowed to receive production traffic.

## `GET /up`

- **Auth**: none (must be reachable before any user session exists, and before the deploy is proven safe).
- **Success response**: `200 OK`, body `ok` (plain text; content is not machine-parsed by kamal-proxy, only the status code is).
  - Returned only once the process has verified it can serve requests — at minimum, that it can reach the configured Postgres database (reusing the existing `storage/psql` connection the app already opens at startup; this endpoint does not open a new connection per request).
- **Failure response**: any non-`2xx` status (e.g., `503 Service Unavailable`) if the dependency check fails. kamal-proxy treats this as "not yet healthy" and keeps routing traffic to the previous version.
- **Timeout/poll behavior**: kamal-proxy polls this endpoint on an interval with a bounded number of attempts (configured in `config/deploy.yml`'s `proxy.healthcheck` block) after the new container starts; exceeding that budget without a `200` is what produces the "deployment failed, traffic never switched" outcome described in User Story 2.
- **Idempotency**: safe to call any number of times; has no side effects.

## Out of scope for this endpoint

- It is not an authenticated endpoint and MUST NOT return any information beyond liveness (no version strings, no stack traces, no internal state) since it is reachable by anything that can reach the proxy.
- It is not a replacement for the existing unauthenticated `GET /` ("hello world") route or the authenticated `GET /hello` route; neither of those is reused for health-check purposes (see `research.md`, Decision 5).
