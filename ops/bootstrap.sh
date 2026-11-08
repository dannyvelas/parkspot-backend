#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

SSH_HOST=45.77.73.20
SSH_PORT=27983
SSH_KEY=~/.ssh/id_ed25519_parkspot-kamal
SSH="ssh -p $SSH_PORT -i $SSH_KEY deploy@$SSH_HOST"

: "${BWS_ACCESS_TOKEN:?BWS_ACCESS_TOKEN must be set}"
BWS_PROJECT_ID=c902bd7c-b27c-4791-bab1-b4e5003dac58

echo "==> kamal setup (Docker, app container, Postgres accessory)"
kamal setup -c ops/deploy.yml

echo "==> Provisioning the restricted parkspot_app Postgres role"
BWS_SECRETS_FILE=$(mktemp)
trap 'rm -f "$BWS_SECRETS_FILE"' EXIT
bws secret list "$BWS_PROJECT_ID" --access-token "$BWS_ACCESS_TOKEN" --output json > "$BWS_SECRETS_FILE"
PGPASSWORD=$(jq -r '.[] | select(.key=="POSTGRES_PASSWORD") | .value' "$BWS_SECRETS_FILE")
APP_DB_PASSWORD=$(jq -r '.[] | select(.key=="POSTGRES_APP_PASSWORD") | .value' "$BWS_SECRETS_FILE")

scp -P "$SSH_PORT" -i "$SSH_KEY" \
  ops/postgres/provision-app-role.sh ops/postgres/provision-app-role.sql \
  "deploy@$SSH_HOST:/tmp/"
$SSH "PGPASSWORD='$PGPASSWORD' APP_DB_PASSWORD='$APP_DB_PASSWORD' bash /tmp/provision-app-role.sh; rm -f /tmp/provision-app-role.sh /tmp/provision-app-role.sql"

echo "==> Installing the backup cron/timer"
if [ -f ops/backup/parkspot-backup.timer ] && [ -f ops/backup/parkspot-backup.service ]; then
  scp -P "$SSH_PORT" -i "$SSH_KEY" \
    ops/backup/backup.sh ops/backup/parkspot-backup.timer ops/backup/parkspot-backup.service \
    "deploy@$SSH_HOST:/tmp/"
  $SSH "sudo mv /tmp/backup.sh /usr/local/bin/parkspot-backup.sh && sudo chmod +x /usr/local/bin/parkspot-backup.sh && \
        sudo mv /tmp/parkspot-backup.timer /tmp/parkspot-backup.service /etc/systemd/system/ && \
        sudo systemctl daemon-reload && sudo systemctl enable --now parkspot-backup.timer"
else
  echo "    ops/backup/parkspot-backup.{timer,service} don't exist yet (US6) -- skipping"
fi

echo "==> Bootstrap complete"
