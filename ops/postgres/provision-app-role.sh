#!/usr/bin/env bash
set -euo pipefail

# Wrapper for provision-app-role.sql. Run this from the Postgres
# accessory's host (or through an SSH tunnel to 127.0.0.1:5432 on it) -
# the accessory only binds to localhost, not the public interface.
#
# Requires:
#   PGPASSWORD      - the postgres superuser's password (to connect)
#   APP_DB_PASSWORD - the password to set for parkspot_app
# Both come from wherever you're keeping production secrets; never
# hardcode them here.

: "${PGPASSWORD:?PGPASSWORD (postgres superuser password) must be set}"
: "${APP_DB_PASSWORD:?APP_DB_PASSWORD must be set}"

psql -h 127.0.0.1 -p 5432 -U postgres -d postgres \
  -v app_password="$APP_DB_PASSWORD" \
  -f "$(dirname "$0")/provision-app-role.sql"
