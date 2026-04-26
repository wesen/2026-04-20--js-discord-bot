---
Title: Discord UI showcase bot k3s deployment design and intern implementation guide
Ticket: DISCORD-BOT-K3S-SHOWCASE-DEPLOY
Status: active
Topics:
    - discord-bot
    - kubernetes
    - k3s
    - vault
    - gitops
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../2026-03-27--hetzner-k3s/gitops/applications/smailnail.yaml
      Note: Argo CD Application pattern
    - Path: ../../../../../../../2026-03-27--hetzner-k3s/gitops/kustomize/smailnail/runtime-secret.yaml
      Note: VaultStaticSecret pattern for runtime secrets
    - Path: README.md
      Note: Runtime overview
    - Path: examples/discord-bots/ui-showcase/index.js
      Note: Selected showcase bot commands and behavior
    - Path: internal/config/config.go
      Note: Credential settings and validation
ExternalSources: []
Summary: Design and implementation guide for deploying the ui-showcase Discord bot to the Hetzner k3s cluster with credentials seeded from .envrc into Vault.
LastUpdated: 2026-04-26T15:58:00-04:00
WhatFor: Use this when implementing or reviewing the k3s deployment of the Discord UI showcase bot.
WhenToUse: Use before creating the GitOps package, Vault policy/role, or running the first Argo CD bootstrap for discord-ui-showcase.
---


# Discord UI showcase bot k3s deployment design and intern implementation guide

## Executive summary

We want to deploy the `ui-showcase` Discord bot from `/home/manuel/code/wesen/2026-04-20--js-discord-bot` to Manuel's Hetzner k3s cluster in `/home/manuel/code/wesen/2026-03-27--hetzner-k3s`. The bot is a Go-hosted Discord gateway process that embeds JavaScript with goja. The JavaScript file `examples/discord-bots/ui-showcase/index.js` owns the Discord behavior, while the Go binary owns credentials, gateway connectivity, slash-command synchronization, and dispatch into JavaScript handlers.

The recommended deployment is a single-replica Kubernetes `Deployment` managed by Argo CD. Runtime Discord credentials should not be committed to Git. They should be copied from this repository's `.envrc` into Vault at `kv/apps/discord-ui-showcase/prod/runtime`, rendered into the cluster by HashiCorp Vault Secrets Operator (`VaultStaticSecret`), and consumed by the pod as environment variables. The implementation should follow the existing k3s patterns used by `smailnail`: `VaultConnection`, `VaultAuth`, `VaultStaticSecret`, a service account bound to a Vault Kubernetes auth role, and a repo-managed Argo CD `Application`.

This ticket includes helper scripts:

- `scripts/01-seed-discord-ui-showcase-vault.sh` seeds the Discord credential keys into Vault without printing values.
- `scripts/02-render-discord-ui-showcase-k3s-manifests.py` renders starter Kubernetes, Argo CD, Vault policy, and Vault role manifests into `var/generated/discord-ui-showcase/`.
- `scripts/03-validate-discord-ui-showcase-deploy.sh` validates Argo status, the rendered Kubernetes secret keys, rollout status, pods, and recent logs.

A first Vault seed attempt was made with `VAULT_ADDR=https://vault.yolo.scapegoat.dev` and the local `~/.vault-token`; Vault was reachable but the local token was invalid, so the secret has not yet been confirmed as written. The operator must repeat the seed with a valid Vault token.

## Problem statement and scope

### Goal

Deploy the showcase Discord bot to the k3s cluster so it stays online without a local terminal session and uses cluster-managed runtime secrets.

The user request is specifically to deploy "this bot (the showcase one)" and to put the credentials currently stored in `.envrc` into Vault. In this repository, "showcase" maps to `examples/discord-bots/ui-showcase/`, not the older `ping` showcase. The `ui-showcase` bot is the comprehensive UI DSL demonstration listed in `examples/discord-bots/README.md` as builder patterns, modal forms, stateful search/review screens, paginated lists, card galleries, confirmations, all select menu types, and alias registration.

### In scope

- Explain the Discord bot runtime architecture.
- Explain the `ui-showcase` bot behavior and command surface.
- Explain the k3s deployment model in this environment.
- Design the Vault path and Kubernetes secret rendering flow.
- Provide concrete GitOps manifests to create in the k3s repository.
- Provide a safe credential-seeding script that reads existing environment variables and writes them into Vault.
- Provide validation steps for a new intern.

### Out of scope

- Rotating the Discord bot token in the Discord Developer Portal.
- Creating a new Discord application.
- Replacing the current runtime with an HTTP interactions-only deployment.
- Adding durable storage to `ui-showcase`; it currently uses in-memory/demo state only.
- Making the deployment highly available. A gateway Discord bot should start as one replica unless the code has explicit multi-instance coordination.

## Current-state evidence

### The repository is a Go-hosted JavaScript Discord bot runtime

The root README describes the core runtime contract: Go owns the Discord gateway/session and embeds a JavaScript runtime, while JavaScript owns bot behavior through `require("discord")` (`README.md:3` and `README.md:7`). It also states the current process model: one selected JavaScript bot per process (`README.md:7`). That matters for Kubernetes because one `Deployment` replica maps to one selected JavaScript bot.

The README's architecture diagram names the important packages (`README.md:203-214`):

```text
discord-bot binary (cmd/discord-bot)
    |
    +-- internal/bot/        Discordgo session wrapper
    +-- internal/config/     Host config (credentials, validation)
    +-- internal/jsdiscord/  Embedded JS runtime + require("discord")
    +-- pkg/framework/       Public embedding API
    +-- pkg/botcli/          Repo-driven multi-bot CLI
```

The runtime data flow is:

```text
Discord gateway
  -> discordgo session
  -> jsdiscord.Host
  -> JavaScript command/event/component/modal handler
  -> normalized response payload
  -> Discord REST/gateway response
```

### The required runtime credentials are known

The README lists the environment variables used by the standalone runtime (`README.md:56-65`):

| Variable | Required by host validation | Purpose |
|---|---:|---|
| `DISCORD_BOT_TOKEN` | yes | Discord bot token used to open the gateway and make REST calls. |
| `DISCORD_APPLICATION_ID` | yes | Discord application/client ID used for command sync. |
| `DISCORD_GUILD_ID` | no | Optional guild-scoped command registration for fast local/single-guild sync. |
| `DISCORD_PUBLIC_KEY` | no | Public key for HTTP interaction verification; not needed for the gateway deployment path. |
| `DISCORD_CLIENT_ID` | no | OAuth/client flows; not needed for basic gateway operation. |
| `DISCORD_CLIENT_SECRET` | no | OAuth/client flows; keep secret if present. |

The code confirms the same settings shape in `internal/config/config.go:11-20`. Host validation currently requires only `DISCORD_BOT_TOKEN` and `DISCORD_APPLICATION_ID` (`internal/config/config.go:31-43`). Optional values can still be stored in Vault so the deployment mirrors `.envrc` and remains future-proof.

### The CLI has flags that map to those credentials

The standalone command declares flags for `bot-token`, `application-id`, `guild-id`, `public-key`, `client-id`, `client-secret`, and `bot-script` in `cmd/discord-bot/commands.go:49-57`. The run command can also sync commands before opening the gateway via `--sync-on-start` (`cmd/discord-bot/commands.go:77-80`). The run path decodes the values, validates them, constructs the bot, optionally syncs commands, and opens the Discord session (`cmd/discord-bot/commands.go:118-135`).

The deployment should prefer environment variables because Kubernetes can inject them from a Secret cleanly. The CLI can read environment variables through the existing Glazed/env behavior used by local runs, and the documented local command already shows passing the environment-derived values as flags (`README.md:45-53`).

### The selected bot is `ui-showcase`

The example bot inventory explicitly lists `ui-showcase/` as the comprehensive UI DSL showcase (`examples/discord-bots/README.md:23-34`). Runtime notes describe its command surface and its in-place component interaction behavior (`examples/discord-bots/README.md:97-115`).

The implementation file documents its commands at the top (`examples/discord-bots/ui-showcase/index.js:1-15`):

- `/demo-message`
- `/demo-form`
- `/demo-search`
- `/demo-review`
- `/demo-confirm`
- `/demo-pager`
- `/demo-cards`
- `/demo-selects`
- `/demo-alias-*`
- `/find`
- `/browse`

The bot configures itself as `name: "ui-showcase"` and category `examples` (`examples/discord-bots/ui-showcase/index.js:42-47`). It logs readiness on the Discord ready event (`examples/discord-bots/ui-showcase/index.js:49-53`) and also responds to the message trigger `!showcase` (`examples/discord-bots/ui-showcase/index.js:55-60`). Its select demos include user, role, channel, and mentionable selects (`examples/discord-bots/ui-showcase/index.js:745-780`), so the bot needs normal guild visibility and permissions appropriate for those interactions.

### The k3s repo already uses Vault Secrets Operator and Argo CD

The k3s runtime-secrets playbook explains that a production rollout spans source repo, Terraform/identity where needed, Vault, GitOps, and cluster reconciliation. Its system boundary explicitly includes Vault runtime values, Kubernetes auth roles/policies, `VaultAuth`, `VaultStaticSecret`, `Deployment`, `Service`, `Ingress`, and Argo CD (`../2026-03-27--hetzner-k3s/docs/app-runtime-secrets-and-identity-provisioning-playbook.md`).

For this bot, the shape is simpler than a web app:

- no public Ingress is needed, because the bot uses the Discord gateway;
- no Service is needed unless future metrics/health endpoints are added;
- no database bootstrap is needed, because `ui-showcase` uses demo in-memory state;
- runtime secrets are still needed for Discord credentials.

The `smailnail` package provides the closest app-secret pattern:

- `runtime-secret.yaml` uses `VaultStaticSecret`, `mount: kv`, `type: kv-v2`, and a path under `apps/<app>/prod/runtime` (`../2026-03-27--hetzner-k3s/gitops/kustomize/smailnail/runtime-secret.yaml:1-16`).
- `vault-auth.yaml` binds the Kubernetes service account to a Vault role of the same app name (`../2026-03-27--hetzner-k3s/gitops/kustomize/smailnail/vault-auth.yaml:1-13`).
- `vault-connection.yaml` points in-cluster VSO traffic to `http://vault.vault.svc.cluster.local:8200` (`../2026-03-27--hetzner-k3s/gitops/kustomize/smailnail/vault-connection.yaml:1-9`).
- `deployment.yaml` consumes rendered secret keys through `secretKeyRef` (`../2026-03-27--hetzner-k3s/gitops/kustomize/smailnail/deployment.yaml:31-65`).
- `kustomization.yaml` collects namespace, service account, Vault resources, secrets, and workload (`../2026-03-27--hetzner-k3s/gitops/kustomize/smailnail/kustomization.yaml:1-18`).
- `gitops/applications/smailnail.yaml` shows the Argo `Application` source/destination/sync-policy pattern (`../2026-03-27--hetzner-k3s/gitops/applications/smailnail.yaml:1-23`).
- Vault policy and role files grant read access to `kv/data/apps/smailnail/prod/*` for the app service account (`../2026-03-27--hetzner-k3s/vault/policies/kubernetes/smailnail.hcl:1-7` and `../2026-03-27--hetzner-k3s/vault/roles/kubernetes/smailnail.json:1-7`).

## Proposed architecture

### High-level system diagram

```text
Local operator workstation
  |
  | source .envrc
  | run scripts/01-seed-discord-ui-showcase-vault.sh
  v
Vault kv-v2
  path: kv/apps/discord-ui-showcase/prod/runtime
  keys: DISCORD_BOT_TOKEN, DISCORD_APPLICATION_ID, DISCORD_GUILD_ID, ...
  |
  | Vault Kubernetes auth role: discord-ui-showcase
  v
Vault Secrets Operator in k3s
  |
  | VaultStaticSecret refreshAfter: 30s
  v
Kubernetes Secret
  name: discord-ui-showcase-runtime
  namespace: discord-ui-showcase
  |
  | env.valueFrom.secretKeyRef
  v
Deployment/Pod
  image: ghcr.io/go-go-golems/discord-bot:sha-<commit>
  command: discord-bot bots --bot-repository /app/examples/discord-bots ui-showcase run --sync-on-start
  |
  v
Discord gateway and Discord application commands
```

### Source repo responsibilities

The Discord bot source repo should own:

- the Go code and tests;
- the JavaScript example bot code;
- the Dockerfile or release image workflow for `discord-bot`;
- documentation that explains how to run the bot locally;
- optionally, a `deploy/gitops-targets.json` if this repo later uses shared CI to open GitOps PRs.

It should not own live secret values. `.envrc` is an operator convenience for local runs. Vault is the cluster secret source of truth.

### k3s GitOps repo responsibilities

The Hetzner k3s repo should own:

- `gitops/kustomize/discord-ui-showcase/namespace.yaml`;
- `gitops/kustomize/discord-ui-showcase/serviceaccount.yaml`;
- `gitops/kustomize/discord-ui-showcase/vault-connection.yaml`;
- `gitops/kustomize/discord-ui-showcase/vault-auth.yaml`;
- `gitops/kustomize/discord-ui-showcase/runtime-secret.yaml`;
- `gitops/kustomize/discord-ui-showcase/deployment.yaml`;
- `gitops/kustomize/discord-ui-showcase/kustomization.yaml`;
- `gitops/applications/discord-ui-showcase.yaml`;
- `vault/policies/kubernetes/discord-ui-showcase.hcl`;
- `vault/roles/kubernetes/discord-ui-showcase.json`.

The ticket helper script `scripts/02-render-discord-ui-showcase-k3s-manifests.py` renders starter versions of these files for review.

### Runtime secret contract

Use this Vault path:

```text
kv/apps/discord-ui-showcase/prod/runtime
```

Use these keys:

```text
DISCORD_BOT_TOKEN        required
DISCORD_APPLICATION_ID   required
DISCORD_GUILD_ID         optional but recommended for scoped command sync
DISCORD_PUBLIC_KEY       optional
DISCORD_CLIENT_ID        optional
DISCORD_CLIENT_SECRET    optional
source                   non-secret provenance marker
```

The Kubernetes Secret should keep the same key names. Keeping the names unchanged avoids accidental translation bugs and makes validation easy: if a new intern knows the local `.envrc` names, they can understand the Kubernetes Secret contract immediately.

### Deployment shape

Start with one replica:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: discord-ui-showcase
spec:
  replicas: 1
  template:
    spec:
      serviceAccountName: discord-ui-showcase
      containers:
        - name: discord-ui-showcase
          image: ghcr.io/go-go-golems/discord-bot:sha-REPLACE_ME
          args:
            - bots
            - --bot-repository
            - /app/examples/discord-bots
            - ui-showcase
            - run
            - --sync-on-start
          env:
            - name: DISCORD_BOT_TOKEN
              valueFrom:
                secretKeyRef:
                  name: discord-ui-showcase-runtime
                  key: DISCORD_BOT_TOKEN
            - name: DISCORD_APPLICATION_ID
              valueFrom:
                secretKeyRef:
                  name: discord-ui-showcase-runtime
                  key: DISCORD_APPLICATION_ID
```

Important detail: the image must contain both the `discord-bot` binary and the `examples/discord-bots/ui-showcase` JavaScript files at the path used by `--bot-repository`. If the existing release/package image contains only the binary, add a deployment image Dockerfile that copies `examples/discord-bots` into `/app/examples/discord-bots`.

### Why no Kubernetes Service or Ingress initially

A Discord gateway bot is an outbound client. It opens a WebSocket connection to Discord and receives events there. The cluster does not need to expose an HTTP endpoint for users to interact with the bot. A Service/Ingress would only be needed if we add:

- Prometheus metrics;
- a health endpoint;
- a debug dashboard;
- HTTP interactions instead of gateway events.

Avoiding an Ingress reduces the deployment surface and avoids unnecessary TLS, DNS, and public-route work.

## API and object references

### Discord bot CLI contract

The documented local run pattern is:

```bash
discord-bot bots ui-showcase run \
  --bot-repository ./examples/discord-bots \
  --bot-token "$DISCORD_BOT_TOKEN" \
  --application-id "$DISCORD_APPLICATION_ID" \
  --guild-id "$DISCORD_GUILD_ID" \
  --sync-on-start
```

In Kubernetes, prefer environment injection and keep flags for bot selection and sync behavior:

```bash
discord-bot bots \
  --bot-repository /app/examples/discord-bots \
  ui-showcase run \
  --sync-on-start
```

### Vault Secrets Operator objects

A `VaultConnection` tells VSO how to reach Vault from inside the cluster:

```yaml
apiVersion: secrets.hashicorp.com/v1beta1
kind: VaultConnection
spec:
  address: http://vault.vault.svc.cluster.local:8200
  skipTLSVerify: true
```

A `VaultAuth` tells VSO which Vault Kubernetes-auth role to use:

```yaml
apiVersion: secrets.hashicorp.com/v1beta1
kind: VaultAuth
spec:
  vaultConnectionRef: vault
  method: kubernetes
  mount: kubernetes
  kubernetes:
    role: discord-ui-showcase
    serviceAccount: discord-ui-showcase
```

A `VaultStaticSecret` maps Vault kv-v2 data into a Kubernetes Secret:

```yaml
apiVersion: secrets.hashicorp.com/v1beta1
kind: VaultStaticSecret
spec:
  vaultAuthRef: discord-ui-showcase
  mount: kv
  type: kv-v2
  path: apps/discord-ui-showcase/prod/runtime
  refreshAfter: 30s
  destination:
    name: discord-ui-showcase-runtime
    create: true
    overwrite: true
```

### Vault policy and role contract

Policy:

```hcl
path "kv/data/apps/discord-ui-showcase/prod/*" {
  capabilities = ["read"]
}

path "kv/metadata/apps/discord-ui-showcase/prod/*" {
  capabilities = ["read", "list"]
}
```

Role:

```json
{
  "bound_service_account_names": ["discord-ui-showcase"],
  "bound_service_account_namespaces": ["discord-ui-showcase"],
  "policies": ["discord-ui-showcase"],
  "ttl": "1h",
  "max_ttl": "4h"
}
```

Provisioning commands:

```bash
cd /home/manuel/code/wesen/2026-03-27--hetzner-k3s
vault policy write discord-ui-showcase vault/policies/kubernetes/discord-ui-showcase.hcl
vault write auth/kubernetes/role/discord-ui-showcase @vault/roles/kubernetes/discord-ui-showcase.json
```

## Implementation plan

### Phase 0: Prepare a valid operator shell

A new intern should start by opening two terminals:

1. Source repo terminal:

```bash
cd /home/manuel/code/wesen/2026-04-20--js-discord-bot
```

2. k3s repo terminal:

```bash
cd /home/manuel/code/wesen/2026-03-27--hetzner-k3s
```

Then confirm required tools:

```bash
command -v go
command -v docker
command -v kubectl
command -v vault
command -v jq
```

Confirm the bot repo variables exist without printing values:

```bash
cd /home/manuel/code/wesen/2026-04-20--js-discord-bot
awk 'match($0,/^[[:space:]]*(export[[:space:]]+)?([A-Za-z_][A-Za-z0-9_]*)=/,m){print m[2]}' .envrc | sort
```

Expected keys from the current `.envrc` are:

```text
DISCORD_APPLICATION_ID
DISCORD_BOT_TOKEN
DISCORD_CLIENT_ID
DISCORD_CLIENT_SECRET
DISCORD_GUILD_ID
DISCORD_PUBLIC_KEY
```

### Phase 1: Seed Vault from `.envrc`

Use the ticket script. It will not print secret values.

```bash
cd /home/manuel/code/wesen/2026-04-20--js-discord-bot
set -a
source ./.envrc
set +a

export VAULT_ADDR=https://vault.yolo.scapegoat.dev
export VAULT_TOKEN='<obtain a valid operator token; do not paste into docs>'

ttmp/2026/04/26/DISCORD-BOT-K3S-SHOWCASE-DEPLOY--deploy-discord-ui-showcase-bot-to-k3s/scripts/01-seed-discord-ui-showcase-vault.sh
```

Validate without exposing values:

```bash
vault kv get -format=json kv/apps/discord-ui-showcase/prod/runtime \
  | jq -r '.data.data | keys[]'
```

Expected keys:

```text
DISCORD_APPLICATION_ID
DISCORD_BOT_TOKEN
DISCORD_CLIENT_ID
DISCORD_CLIENT_SECRET
DISCORD_GUILD_ID
DISCORD_PUBLIC_KEY
source
```

If `vault kv get` returns `permission denied` or `invalid token`, stop and obtain a valid Vault token. Do not work around this by committing secrets into Kubernetes YAML.

### Phase 2: Build and publish an image that contains examples

The deployment assumes a container image containing:

- `/usr/local/bin/discord-bot` or equivalent entrypoint binary;
- `/app/examples/discord-bots/ui-showcase/index.js`;
- `/app/examples/discord-bots/ui-showcase/lib/...`;
- any JavaScript modules loaded relative to the bot script.

Pseudocode Dockerfile:

```dockerfile
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o /out/discord-bot ./cmd/discord-bot

FROM debian:bookworm-slim
RUN useradd -r -u 10001 appuser
WORKDIR /app
COPY --from=build /out/discord-bot /usr/local/bin/discord-bot
COPY examples/discord-bots /app/examples/discord-bots
USER appuser
ENTRYPOINT ["discord-bot"]
```

Publish as an immutable SHA tag, for example:

```text
ghcr.io/go-go-golems/discord-bot:sha-<short-sha>
```

If this repository already has a release workflow that publishes images, use that. If not, add a minimal GitHub Actions workflow following the k3s public-repo GHCR playbook.

### Phase 3: Generate starter manifests

From the ticket workspace:

```bash
cd /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/26/DISCORD-BOT-K3S-SHOWCASE-DEPLOY--deploy-discord-ui-showcase-bot-to-k3s
./scripts/02-render-discord-ui-showcase-k3s-manifests.py
```

The generated files land in:

```text
var/generated/discord-ui-showcase/
```

Review them, then copy into the k3s repo:

```bash
cd /home/manuel/code/wesen/2026-03-27--hetzner-k3s
mkdir -p gitops/kustomize/discord-ui-showcase
cp /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/26/DISCORD-BOT-K3S-SHOWCASE-DEPLOY--deploy-discord-ui-showcase-bot-to-k3s/var/generated/discord-ui-showcase/{namespace.yaml,serviceaccount.yaml,vault-connection.yaml,vault-auth.yaml,runtime-secret.yaml,deployment.yaml,kustomization.yaml} \
  gitops/kustomize/discord-ui-showcase/
cp /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/26/DISCORD-BOT-K3S-SHOWCASE-DEPLOY--deploy-discord-ui-showcase-bot-to-k3s/var/generated/discord-ui-showcase/application.yaml \
  gitops/applications/discord-ui-showcase.yaml
cp /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/26/DISCORD-BOT-K3S-SHOWCASE-DEPLOY--deploy-discord-ui-showcase-bot-to-k3s/var/generated/discord-ui-showcase/vault-policy.hcl \
  vault/policies/kubernetes/discord-ui-showcase.hcl
cp /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/26/DISCORD-BOT-K3S-SHOWCASE-DEPLOY--deploy-discord-ui-showcase-bot-to-k3s/var/generated/discord-ui-showcase/vault-role.json \
  vault/roles/kubernetes/discord-ui-showcase.json
```

Then replace `sha-REPLACE_ME` in `deployment.yaml` with the real published image tag.

### Phase 4: Validate render and Vault auth files

Run:

```bash
cd /home/manuel/code/wesen/2026-03-27--hetzner-k3s
kubectl kustomize gitops/kustomize/discord-ui-showcase >/tmp/discord-ui-showcase.rendered.yaml
```

Check the rendered YAML contains exactly the expected namespace and resource names:

```bash
rg -n 'kind:|name: discord-ui-showcase|discord-ui-showcase-runtime|apps/discord-ui-showcase/prod/runtime' /tmp/discord-ui-showcase.rendered.yaml
```

Provision Vault policy and role:

```bash
vault policy write discord-ui-showcase vault/policies/kubernetes/discord-ui-showcase.hcl
vault write auth/kubernetes/role/discord-ui-showcase @vault/roles/kubernetes/discord-ui-showcase.json
```

### Phase 5: Merge or apply GitOps package

Create a normal GitOps PR in the k3s repo with:

- app manifests under `gitops/kustomize/discord-ui-showcase/`;
- Argo application under `gitops/applications/discord-ui-showcase.yaml`;
- Vault policy and role under `vault/`.

After merge, remember the first-time Argo bootstrap rule: this repo does not auto-create new Argo `Application` objects just because a file exists. Apply the Application once:

```bash
cd /home/manuel/code/wesen/2026-03-27--hetzner-k3s
export KUBECONFIG=$PWD/kubeconfig-91.98.46.169.yaml  # or the Tailscale kubeconfig for this cluster
kubectl apply -f gitops/applications/discord-ui-showcase.yaml
kubectl -n argocd annotate application discord-ui-showcase argocd.argoproj.io/refresh=hard --overwrite
```

### Phase 6: Validate cluster rollout

Run the ticket validation script:

```bash
cd /home/manuel/code/wesen/2026-04-20--js-discord-bot
export KUBECONFIG=/home/manuel/code/wesen/2026-03-27--hetzner-k3s/kubeconfig-91.98.46.169.yaml
ttmp/2026/04/26/DISCORD-BOT-K3S-SHOWCASE-DEPLOY--deploy-discord-ui-showcase-bot-to-k3s/scripts/03-validate-discord-ui-showcase-deploy.sh
```

Manual validation commands:

```bash
kubectl -n argocd get application discord-ui-showcase
kubectl -n discord-ui-showcase get vaultauth,vaultstaticsecret,secret,pods
kubectl -n discord-ui-showcase describe vaultstaticsecret discord-ui-showcase-runtime
kubectl -n discord-ui-showcase rollout status deployment/discord-ui-showcase --timeout=180s
kubectl -n discord-ui-showcase logs deployment/discord-ui-showcase --tail=100
```

Discord validation:

1. In the target guild, run `/demo-message`.
2. Click the primary/success/danger buttons.
3. Run `/demo-form` and submit the modal.
4. Run `/demo-search` or `/find` with a query such as `discord`.
5. Send `!showcase` in a channel where the bot can read messages.

If slash commands do not appear, check whether `DISCORD_GUILD_ID` was provided and whether `--sync-on-start` succeeded in logs.

## Pseudocode for the end-to-end rollout

```text
function deploy_discord_ui_showcase():
    assert local_repo == discord_bot_repo
    env = load_envrc_without_printing_values(".envrc")
    require env.DISCORD_BOT_TOKEN
    require env.DISCORD_APPLICATION_ID

    vault = login_to_vault(operator_token)
    vault.kv_put(
        mount="kv",
        path="apps/discord-ui-showcase/prod/runtime",
        values={
            "DISCORD_BOT_TOKEN": env.DISCORD_BOT_TOKEN,
            "DISCORD_APPLICATION_ID": env.DISCORD_APPLICATION_ID,
            "DISCORD_GUILD_ID": env.optional("DISCORD_GUILD_ID"),
            "DISCORD_PUBLIC_KEY": env.optional("DISCORD_PUBLIC_KEY"),
            "DISCORD_CLIENT_ID": env.optional("DISCORD_CLIENT_ID"),
            "DISCORD_CLIENT_SECRET": env.optional("DISCORD_CLIENT_SECRET"),
            "source": "discord-bot .envrc",
        },
    )

    image = build_and_publish_image(repo="discord-bot", tag="sha-<commit>")
    assert image.contains("/usr/local/bin/discord-bot")
    assert image.contains("/app/examples/discord-bots/ui-showcase/index.js")

    k3s_repo.write("gitops/kustomize/discord-ui-showcase/*")
    k3s_repo.write("gitops/applications/discord-ui-showcase.yaml")
    k3s_repo.write("vault/policies/kubernetes/discord-ui-showcase.hcl")
    k3s_repo.write("vault/roles/kubernetes/discord-ui-showcase.json")

    run("kubectl kustomize gitops/kustomize/discord-ui-showcase")
    run("vault policy write discord-ui-showcase vault/policies/kubernetes/discord-ui-showcase.hcl")
    run("vault write auth/kubernetes/role/discord-ui-showcase @vault/roles/kubernetes/discord-ui-showcase.json")

    merge_gitops_pr()
    run("kubectl apply -f gitops/applications/discord-ui-showcase.yaml")
    run("kubectl -n argocd annotate application discord-ui-showcase argocd.argoproj.io/refresh=hard --overwrite")
    run("kubectl -n discord-ui-showcase rollout status deployment/discord-ui-showcase")
```

## Risks, alternatives, and open questions

### Risk: invalid or expired Vault token

Observed during this ticket: `vault status` succeeded against `https://vault.yolo.scapegoat.dev`, but `vault kv put` failed with `permission denied` and `invalid token` using the local `~/.vault-token`. The fix is to use a valid operator token or complete the normal Vault login path. Do not print the token and do not commit credentials as a workaround.

### Risk: image does not contain JavaScript examples

A binary-only release image will fail because `bots --bot-repository /app/examples/discord-bots ui-showcase run` needs the JavaScript files. Validate with:

```bash
docker run --rm --entrypoint sh ghcr.io/go-go-golems/discord-bot:sha-<tag> -c 'ls -la /app/examples/discord-bots/ui-showcase && discord-bot bots --bot-repository /app/examples/discord-bots list'
```

### Risk: two replicas duplicate gateway behavior

Discord gateway bots can often connect multiple sessions, but without explicit sharding or leader election, two pods can duplicate event handling or fight over command sync. Start with `replicas: 1`.

### Risk: command sync scope surprise

If `DISCORD_GUILD_ID` is present, command sync is guild-scoped and should appear quickly in that guild. If it is absent, global application commands can take longer to appear. The intern should know this before declaring the deployment broken.

### Alternative: direct `run --bot-script`

Instead of repository scanning, the pod could run:

```bash
discord-bot run --bot-script /app/examples/discord-bots/ui-showcase/index.js --sync-on-start
```

This is simpler, but the named bot runner path is better here because the repo already documents and tests `bots ui-showcase run` and the selected bot is discoverable in `bots list`/`bots help`.

### Alternative: Kubernetes Secret created manually

A one-off `kubectl create secret generic` would be faster but weaker. It bypasses Vault, makes rotation less consistent, and conflicts with the existing k3s secret-management pattern. Use Vault and VSO.

## Testing strategy

### Local source tests

Before publishing the image:

```bash
cd /home/manuel/code/wesen/2026-04-20--js-discord-bot
make test
make build
GOWORK=off go run ./cmd/discord-bot --bot-repository ./examples/discord-bots bots help ui-showcase --output json
```

### Manifest tests

Before applying Argo:

```bash
cd /home/manuel/code/wesen/2026-03-27--hetzner-k3s
kubectl kustomize gitops/kustomize/discord-ui-showcase >/tmp/discord-ui-showcase.yaml
kubectl apply --dry-run=server -f /tmp/discord-ui-showcase.yaml
```

### Secret tests

After Vault policy/role and Argo sync:

```bash
kubectl -n discord-ui-showcase get secret discord-ui-showcase-runtime -o jsonpath='{range $k,$v := .data}{printf "%s\n" $k}{end}'
kubectl -n discord-ui-showcase describe vaultstaticsecret discord-ui-showcase-runtime
```

Never decode or paste the secret values into docs or chat.

### Runtime tests

Watch logs:

```bash
kubectl -n discord-ui-showcase logs deployment/discord-ui-showcase -f
```

Then test the Discord UX:

- `/demo-message` renders an ephemeral message with buttons and a select.
- `/demo-form` opens a modal and accepts submission.
- `/demo-search query:discord` returns a paginated search screen.
- `/demo-selects` renders string/user/role/channel/mentionable select examples.
- `!showcase` returns the online/help message.

## File references

### Discord bot repo

- `/home/manuel/code/wesen/2026-04-20--js-discord-bot/README.md` — runtime overview, credentials, and architecture.
- `/home/manuel/code/wesen/2026-04-20--js-discord-bot/.envrc` — local credential source; do not commit values elsewhere.
- `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/config/config.go` — credential struct and validation.
- `/home/manuel/code/wesen/2026-04-20--js-discord-bot/cmd/discord-bot/commands.go` — CLI flags and run flow.
- `/home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/README.md` — example bot inventory and runtime notes.
- `/home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/ui-showcase/index.js` — selected showcase bot behavior.
- `/home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/ui-showcase/lib/ui/index.js` — UI DSL re-export from Go-native `require("ui")` plus JS flow helpers.

### k3s repo

- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/docs/app-runtime-secrets-and-identity-provisioning-playbook.md` — cluster-side secret and identity rollout model.
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/docs/source-app-deployment-infrastructure-playbook.md` — source repo to GHCR to GitOps to Argo model.
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/docs/public-repo-ghcr-argocd-deployment-playbook.md` — public GHCR image deployment pattern.
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/docs/argocd-app-setup.md` — Argo CD `Application` setup and first-time bootstrap rule.
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/smailnail/` — closest existing Vault-backed app package pattern.
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/applications/smailnail.yaml` — Argo Application pattern.
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/vault/policies/kubernetes/smailnail.hcl` — Vault policy pattern.
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/vault/roles/kubernetes/smailnail.json` — Vault Kubernetes auth role pattern.

### Ticket helper files

- `/home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/26/DISCORD-BOT-K3S-SHOWCASE-DEPLOY--deploy-discord-ui-showcase-bot-to-k3s/scripts/01-seed-discord-ui-showcase-vault.sh`
- `/home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/26/DISCORD-BOT-K3S-SHOWCASE-DEPLOY--deploy-discord-ui-showcase-bot-to-k3s/scripts/02-render-discord-ui-showcase-k3s-manifests.py`
- `/home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/26/DISCORD-BOT-K3S-SHOWCASE-DEPLOY--deploy-discord-ui-showcase-bot-to-k3s/scripts/03-validate-discord-ui-showcase-deploy.sh`
- `/home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/26/DISCORD-BOT-K3S-SHOWCASE-DEPLOY--deploy-discord-ui-showcase-bot-to-k3s/var/generated/discord-ui-showcase/`
