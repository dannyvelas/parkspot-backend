#!/usr/bin/env bash
set -euo pipefail

# Wrapper for provision-app-role.sql.  psql lives inside the
# "parkspot-backend-postgres" container, not on the host itself.

: "${PGPASSWORD:?PGPASSWORD (postgres superuser password) must be set}"
: "${POSTGRES_APP_PASSWORD:?POSTGRES_APP_PASSWORD must be set}"

CONTAINER=${POSTGRES_CONTAINER:-parkspot-backend-postgres}

docker exec -i -e PGPASSWORD="$PGPASSWORD" "$CONTAINER" \
  psql -U postgres -d parkspot -v app_password="$POSTGRES_APP_PASSWORD" -f - \
  < "$(dirname "$0")/provision-app-role.sql"
