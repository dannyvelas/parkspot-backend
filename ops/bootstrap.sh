#!/usr/bin/env bash
set -euo pipefail

# we use a lot of paths here where we assume this script is being
# run from the root of the repo. e.g. further down we reference "ops/deploy.yml".
# this line makes it so that these lines are referenced from the root of the repo
# instead of whatever arbitrary location the user was in when they ran this script
# dirname $0 == "ops". the parent directory of that is the root of the repo
cd "$(dirname "$0")/.."

SSH_HOST=45.77.73.20
SSH_PORT=27983
SSH_KEY=~/.ssh/id_ed25519_parkspot-kamal
SSH="ssh -p $SSH_PORT -i $SSH_KEY deploy@$SSH_HOST"

# ":" is a bash "no-op" command. what we really want to do is a one-liner where we
# check if BWS_ACCESS_TOKEN is set and error out if not. but if the entire line was
# this: "${BWS_ACCESS_TOKEN:?BWS_ACCESS_TOKEN must be set}", then bash would try to
# expand that to a command, which it's not. so we use ":" to act as the command which
# does nothing.
: "${BWS_ACCESS_TOKEN:?BWS_ACCESS_TOKEN must be set}"
: "${KAMAL_REGISTRY_PASSWORD:?KAMAL_REGISTRY_PASSWORD must be set}"
BWS_PROJECT_ID=c902bd7c-b27c-4791-bab1-b4e5003dac58

echo "==> kamal setup (Docker, app container, Postgres accessory, backup accessory)"
kamal setup -c ops/deploy.yml

echo "==> Provisioning the restricted parkspot_app Postgres role"
BWS_SECRETS_FILE=$(mktemp)
PROVISION_SECRETS_FILE=$(mktemp)

# regardless of how this script ends, run rm -r "$BWS_SECRETS_FILE right before
trap 'rm -f "$BWS_SECRETS_FILE" "$PROVISION_SECRETS_FILE"' EXIT

bws secret list "$BWS_PROJECT_ID" --access-token "$BWS_ACCESS_TOKEN" --output json > "$BWS_SECRETS_FILE"
# named PGPASSWORD, not POSTGRES_PASSWORD, because psql and pg_dump expect that exact env name
PGPASSWORD=$(jq -r '.[] | select(.key=="POSTGRES_PASSWORD") | .value' "$BWS_SECRETS_FILE")
POSTGRES_APP_PASSWORD=$(jq -r '.[] | select(.key=="POSTGRES_APP_PASSWORD") | .value' "$BWS_SECRETS_FILE")

chmod 600 "$PROVISION_SECRETS_FILE"
printf 'PGPASSWORD=%s\nPOSTGRES_APP_PASSWORD=%s\n' "$PGPASSWORD" "$POSTGRES_APP_PASSWORD" > "$PROVISION_SECRETS_FILE"

scp -P "$SSH_PORT" -i "$SSH_KEY" \
  ops/postgres/provision-app-role.sh ops/postgres/provision-app-role.sql "$PROVISION_SECRETS_FILE" \
  "deploy@$SSH_HOST:/tmp/"
$SSH "mv /tmp/$(basename "$PROVISION_SECRETS_FILE") /tmp/provision-secrets.env && chmod 600 /tmp/provision-secrets.env && \
      set -a && . /tmp/provision-secrets.env && set +a && bash /tmp/provision-app-role.sh; \
      rm -f /tmp/provision-app-role.sh /tmp/provision-app-role.sql /tmp/provision-secrets.env"

echo "==> Bootstrap complete"
