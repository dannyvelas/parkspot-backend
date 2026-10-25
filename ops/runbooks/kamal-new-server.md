# Deploying parkspot-backend to a new server with Kamal

A step-by-step runbook for taking a freshly provisioned server to a working `kamal deploy` for this repo, with secrets resolved from Bitwarden Secrets Manager at deploy time via `.kamal/secrets`.

This covers `kamal setup` + `kamal deploy` run from a laptop.

| | Where |
|---|---|
| Service | `parkspot-backend` |
| Image | `ghcr.io/dannyvelas/parkspot-backend` |
| Domain | `api.park-spot.co` |
| App port | `5000` |
| Host | `servers.web.hosts` |
| SSH port / key | `ssh.port`, `ssh.keys` |
| Deploy user | `ssh.user` |
| Database | `postgres:17.6` accessory, container `parkspot-backend-postgres` |

The commands below use these shell variables. Set them once per session, from `config/deploy.yml` and from `terraform output` in the [infra](https://github.com/dannyvelas/infra) repo:

```fish
set ip        (yq -r '.servers.web.hosts[0]' config/deploy.yml)
set ssh_port  (yq -r '.ssh.port'             config/deploy.yml)
set ssh_key   (yq -r '.ssh.keys[0]'          config/deploy.yml | string replace '~' $HOME)
set admin_user (terraform output -raw admin_username)   # in the infra repo
```

## What you need

- `kamal` installed (`gem install kamal`; currently 2.10.1)
- `bws` and `jq` on `PATH` — `.kamal/secrets` shells out to both
- `yq` on `PATH`, for reading connection details out of `config/deploy.yml`
- A `BWS_ACCESS_TOKEN` set in your env for the Bitwarden Secrets Manager project `c902bd7c-b27c-4791-bab1-b4e5003dac58`. Only having "read" permissions on this is fine.
- A `KAMAL_REGISTRY_PASSWORD` set in your env for `kamal setup`. You can get this by generating a new personal access token on Github. The scope needed is `write:packages`. This allows Kamal to push and pull `ghcr.io/dannyvelas/parkspot-backend`.

## Step 1: Provision the server

Server provisioning (the instance, a dedicated SSH key, a deploy user with passwordless sudo, hardened SSH, a firewall, automatic security updates, docker) is automated by the [infra](https://github.com/dannyvelas/infra) Terraform project. Follow its README; don't provision by hand.

That project generates a **new key pair per server** and never reuses one, so the key in `ssh.keys` is specific to the current box and changes with it. After `terraform apply`, get the connection details with `terraform output` (`ip_address`, `admin_username`, `ssh_port`), and copy the printed SSH key passphrase and console-recovery password into Bitwarden before closing that terminal — they are shown exactly once.

When provisioning a replacement server, update `servers.web.hosts`, `accessories.postgres.host`, `ssh.port`, and `ssh.keys` in `config/deploy.yml` first — the rest of this runbook reads from there.

## Step 2: Point DNS at the server

`proxy.ssl: true` makes `kamal-proxy` request a Let's Encrypt certificate for `api.park-spot.co` on boot, which only works once that name resolves to this server and ports 80/443 are reachable. Update the A record before deploying, and confirm it has propagated:

```sh
dig +short api.park-spot.co
```

## Step 3: Review `config/deploy.yml`

The config is checked in, so there is nothing to write from scratch — but these are the fields that cause real failures, and the ones to change when moving to a different server:

- **`servers.web.hosts`** and **`accessories.postgres.host`** must *both* be updated when the server changes. Kamal's host list is the union of `servers:` and every `accessories.*.host`, and `kamal setup` bootstraps every host in that union — so a stale accessory host silently makes Kamal connect to the old box and boot Postgres there instead of here.
- **`proxy.app_port: 5000`** must match the port the app actually listens on. Left out, `kamal-proxy` health-checks port 80, which passes Docker's own health check but fails every real request. **`builder.arch: amd64`** must be set explicitly. Without it, `kamal setup` fails immediately with `Kamal::ConfigurationError: builder: Builder arch not set`, before it even reaches the server. **`ssh.port`** must be the hardened port set by the infra project, not `22`, and **`ssh.keys`** that server's dedicated key, not your default `~/.ssh/id_ed25519`.
- **`env.secret`** must name exactly the keys that `.kamal/secrets` defines, which in turn must exist in Bitwarden. A name listed here but missing downstream means the app boots with that variable silently empty rather than erroring, so the three lists have to be kept in sync by hand.

## Step 4: Run `kamal setup`

```fish
kamal setup
```

This installs Docker's buildx builder locally if needed, builds and pushes the image, boots `kamal-proxy`, boots the Postgres accessory, and starts the app container.

## Step 5: Subsequent deploys

```sh
kamal deploy
```

Both `KAMAL_REGISTRY_PASSWORD` and `BWS_ACCESS_TOKEN` still need to be set in whatever shell runs this. Nothing needs regenerating when a value changes in Bitwarden — `.kamal/secrets` is a live lookup, re-resolved on every deploy.
