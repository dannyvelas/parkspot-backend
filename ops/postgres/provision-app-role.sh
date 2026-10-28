#!/usr/bin/env bash
set -euo pipefail

# Wrapper for provision-app-role.sql. Run this on the Postgres
# accessory's host, as a user that can run `docker exec` against it
# (e.g. a member of the docker group) - psql lives inside the
# "parkspot-backend-postgres" container, not on the host itself.
#
# Requires:
#   PGPASSWORD      - the postgres superuser's password (to connect)
#   APP_DB_PASSWORD - the password to set for parkspot_app
# Both come from wherever you're keeping production secrets; never
# hardcode them here.

: "${PGPASSWORD:?PGPASSWORD (postgres superuser password) must be set}"
: "${APP_DB_PASSWORD:?APP_DB_PASSWORD must be set}"

CONTAINER=${POSTGRES_CONTAINER:-parkspot-backend-postgres}

docker exec -i -e PGPASSWORD="$PGPASSWORD" "$CONTAINER" \
  psql -U postgres -d parkspot -v app_password="$APP_DB_PASSWORD" -f - \
  < "$(dirname "$0")/provision-app-role.sql"
