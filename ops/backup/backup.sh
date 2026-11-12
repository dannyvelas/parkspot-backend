#!/usr/bin/env bash
set -euo pipefail

: "${POSTGRES_APP_PASSWORD:?POSTGRES_APP_PASSWORD must be set}"
: "${VULTR_OBJECT_STORAGE_ACCESS_KEY:?VULTR_OBJECT_STORAGE_ACCESS_KEY must be set}"
: "${VULTR_OBJECT_STORAGE_SECRET_KEY:?VULTR_OBJECT_STORAGE_SECRET_KEY must be set}"

POSTGRES_HOST=parkspot-backend-postgres
S3_ENDPOINT=https://atl2.vultrobjects.com
S3_BUCKET=parkspot-backups

export AWS_ACCESS_KEY_ID="$VULTR_OBJECT_STORAGE_ACCESS_KEY"
export AWS_SECRET_ACCESS_KEY="$VULTR_OBJECT_STORAGE_SECRET_KEY"
export AWS_DEFAULT_REGION=us-east-1
export PGPASSWORD="$POSTGRES_APP_PASSWORD"

# check if bucket already exists. if not, create it. this should only execute the first time this script is run
if ! aws --endpoint-url "$S3_ENDPOINT" s3api head-bucket --bucket "$S3_BUCKET" 2>/dev/null; then
  aws --endpoint-url "$S3_ENDPOINT" s3 mb "s3://$S3_BUCKET"
fi

DUMP_FILE=$(mktemp -t parkspot-backup-XXXXXX.dump)
trap 'rm -f "$DUMP_FILE"' EXIT

# kamal created a docker network via `docker network create`. in a docker network,
# docker containers can resolve eachother by name. this is because docker runs a
# DNS server. whenever a container looks up the name of another container, the
# the DNS server will respond with an IP address
pg_dump --format=custom -h "$POSTGRES_HOST" -U parkspot_app -d parkspot > "$DUMP_FILE"

OBJECT_KEY="parkspot-$(date -u +%Y%m%dT%H%M%SZ).dump"
aws --endpoint-url "$S3_ENDPOINT" s3 cp "$DUMP_FILE" "s3://$S3_BUCKET/$OBJECT_KEY"
