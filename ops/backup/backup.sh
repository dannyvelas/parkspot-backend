#!/usr/bin/env bash
set -euo pipefail

: "${POSTGRES_APP_PASSWORD:?POSTGRES_APP_PASSWORD must be set}"
: "${VULTR_OBJECT_STORAGE_ACCESS_KEY:?VULTR_OBJECT_STORAGE_ACCESS_KEY must be set}"
: "${VULTR_OBJECT_STORAGE_SECRET_KEY:?VULTR_OBJECT_STORAGE_SECRET_KEY must be set}"

POSTGRES_CONTAINER=parkspot-backend-postgres
S3_ENDPOINT=https://atl2.vultrobjects.com
S3_BUCKET=parkspot-backups

export AWS_ACCESS_KEY_ID="$VULTR_OBJECT_STORAGE_ACCESS_KEY"
export AWS_SECRET_ACCESS_KEY="$VULTR_OBJECT_STORAGE_SECRET_KEY"
export AWS_DEFAULT_REGION=us-east-1

if ! aws --endpoint-url "$S3_ENDPOINT" s3api head-bucket --bucket "$S3_BUCKET" 2>/dev/null; then
  aws --endpoint-url "$S3_ENDPOINT" s3 mb "s3://$S3_BUCKET"
fi

DUMP_FILE=$(mktemp -t parkspot-backup-XXXXXX.dump)
trap 'rm -f "$DUMP_FILE"' EXIT

docker exec -i -e PGPASSWORD="$POSTGRES_APP_PASSWORD" "$POSTGRES_CONTAINER" \
  pg_dump --format=custom -U parkspot_app -d parkspot > "$DUMP_FILE"

OBJECT_KEY="parkspot-$(date -u +%Y%m%dT%H%M%SZ).dump"
aws --endpoint-url "$S3_ENDPOINT" s3 cp "$DUMP_FILE" "s3://$S3_BUCKET/$OBJECT_KEY"
