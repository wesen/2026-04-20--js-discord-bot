---
Title: Investigation diary
Ticket: DISCORD-BOT-K3S-SHOWCASE-DEPLOY
Status: active
Topics:
    - discord-bot
    - kubernetes
    - k3s
    - vault
    - gitops
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/04/26/DISCORD-BOT-K3S-SHOWCASE-DEPLOY--deploy-discord-ui-showcase-bot-to-k3s/scripts/01-seed-discord-ui-showcase-vault.sh
      Note: Safe Vault seed script from .envrc
    - Path: ttmp/2026/04/26/DISCORD-BOT-K3S-SHOWCASE-DEPLOY--deploy-discord-ui-showcase-bot-to-k3s/scripts/02-render-discord-ui-showcase-k3s-manifests.py
      Note: Starter GitOps/Vault manifest generator
    - Path: ttmp/2026/04/26/DISCORD-BOT-K3S-SHOWCASE-DEPLOY--deploy-discord-ui-showcase-bot-to-k3s/scripts/03-validate-discord-ui-showcase-deploy.sh
      Note: Post-deploy validation script
ExternalSources: []
Summary: Chronological investigation diary for the Discord UI showcase bot k3s deployment guide.
LastUpdated: 2026-04-26T15:58:00-04:00
WhatFor: Record what was investigated, generated, and attempted for the deployment ticket.
WhenToUse: Read before resuming implementation of the discord-ui-showcase k3s deployment.
---


# Diary

## Goal

Capture the investigation and documentation work for deploying the Discord `ui-showcase` bot to the Hetzner k3s cluster, including the Vault credential-seeding attempt and the generated helper scripts.

## Step 1: Create ticket and map the source and cluster patterns

I created a new docmgr ticket for the deployment work and inspected both the Discord bot repository and the Hetzner k3s repository. The main goal was to produce an intern-friendly design and implementation guide that explains the whole system rather than just listing commands.

The key finding is that the Discord bot repo already has a named `ui-showcase` bot under `examples/discord-bots/ui-showcase`, and the k3s repo already has a reusable Vault Secrets Operator plus Argo CD deployment pattern in the `smailnail` package.

### Prompt Context

**User prompt (verbatim):** "Create a new docmgr ticket to deploy this bot (the showcase one) to my k3s cluster (see ../2026-03-27--hetzner-k3s), by putting the credentials in .envrc into the vault. 

reate a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet  
point sand pseudocode and diagrams and api references and file                                                                                                                                              
  references.                                                                                                                                                                                               
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable.                                                                                                                  
                                                                                                                                                                                                            
You can write scripts and all kinds of other things in the scripts/ folder of the ticket."

**Assistant interpretation:** Create a docmgr ticket containing a detailed deployment guide for the `ui-showcase` Discord bot, include scripts under the ticket workspace, seed or prepare to seed credentials from `.envrc` into Vault, validate the docs, and upload the bundle to reMarkable.

**Inferred user intent:** The user wants a complete handoff package that a new intern can use to safely implement the k3s deployment without leaking secrets or misunderstanding the repo/cluster boundaries.

**Commit (code):** N/A — no git commit was requested or made.

### What I did

- Ran `docmgr ticket create-ticket --ticket DISCORD-BOT-K3S-SHOWCASE-DEPLOY --title "Deploy Discord UI showcase bot to k3s" --topics discord-bot,kubernetes,k3s,vault,gitops`.
- Added a design doc with `docmgr doc add --ticket DISCORD-BOT-K3S-SHOWCASE-DEPLOY --doc-type design-doc --title "Discord UI showcase bot k3s deployment design and intern implementation guide"`.
- Added this diary with `docmgr doc add --ticket DISCORD-BOT-K3S-SHOWCASE-DEPLOY --doc-type reference --title "Investigation diary"`.
- Inspected `README.md`, `examples/discord-bots/README.md`, `internal/config/config.go`, `cmd/discord-bot/commands.go`, and `examples/discord-bots/ui-showcase/index.js`.
- Inspected k3s docs and examples including `docs/app-runtime-secrets-and-identity-provisioning-playbook.md`, `docs/source-app-deployment-infrastructure-playbook.md`, `docs/public-repo-ghcr-argocd-deployment-playbook.md`, `docs/argocd-app-setup.md`, and `gitops/kustomize/smailnail/*`.
- Extracted `.envrc` variable names without printing values:
  - `DISCORD_APPLICATION_ID`
  - `DISCORD_BOT_TOKEN`
  - `DISCORD_CLIENT_ID`
  - `DISCORD_CLIENT_SECRET`
  - `DISCORD_GUILD_ID`
  - `DISCORD_PUBLIC_KEY`

### Why

- The deployment guide needed evidence-backed file references.
- The intern guide needed to explain both the application runtime and the platform deployment pattern.
- Reading variable names from `.envrc` was necessary to design the Vault secret contract, but printing values would have been unsafe.

### What worked

- The docmgr ticket and documents were created successfully.
- The source repo clearly identifies `ui-showcase` as the showcase bot.
- The k3s repo has mature examples for Argo CD, Vault Secrets Operator, Vault policies, and Vault Kubernetes auth roles.

### What didn't work

- N/A for this step.

### What I learned

- The Discord bot runtime is a Go gateway host with embedded JavaScript, not a conventional HTTP server.
- `ui-showcase` should start as a single-replica outbound gateway workload with no Service or Ingress.
- The closest k3s pattern is `smailnail`, but this bot is simpler because it has no database and no browser OIDC flow.

### What was tricky to build

- The user said "showcase one," and the repo has both a `ping` API showcase and a `ui-showcase` UI DSL showcase. The explicit files and runtime notes showed that `ui-showcase` is the more comprehensive showcase requested here.
- The guide had to include `.envrc` credentials without exposing values. I used `awk` to list variable names only.

### What warrants a second pair of eyes

- Confirm that `ui-showcase` is indeed the desired showcase bot rather than `ping`.
- Confirm the final GHCR image path and whether the image contains `examples/discord-bots`.

### What should be done in the future

- Use a valid Vault operator login and rerun the seed script.
- Copy generated manifests into the k3s repo only after image publishing is confirmed.

### Code review instructions

- Start with the design doc in `design-doc/01-discord-ui-showcase-bot-k3s-deployment-design-and-intern-implementation-guide.md`.
- Compare the generated manifests in `var/generated/discord-ui-showcase/` against `../2026-03-27--hetzner-k3s/gitops/kustomize/smailnail/`.
- Validate with `kubectl kustomize` in the k3s repo before applying anything.

### Technical details

Important commands run:

```bash
rg -n "showcase|SHOWCASE|Discord.*bot|DISCORD" README.md examples cmd internal pkg .envrc.example -S
awk 'match($0,/^[[:space:]]*(export[[:space:]]+)?([A-Za-z_][A-Za-z0-9_]*)=/,m){print m[2]}' .envrc | sort
nl -ba README.md | sed -n '1,90p;175,215p'
nl -ba examples/discord-bots/README.md | sed -n '1,45p;65,105p;113,118p'
nl -ba internal/config/config.go | sed -n '1,80p'
nl -ba examples/discord-bots/ui-showcase/index.js | sed -n '1,70p;105,150p;745,790p'
```

## Step 2: Add ticket helper scripts and generated starter manifests

I added scripts under the ticket `scripts/` directory because the user explicitly allowed scripts and because a repeatable, non-secret-printing workflow is safer than prose-only instructions. The scripts seed Vault, render starter manifests, and validate the deployment after Argo sync.

The render script was executed once and wrote generated starter files under `var/generated/discord-ui-showcase/` for review and copy into the k3s repo.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Provide executable scaffolding to make the guide actionable and safe for an intern.

**Inferred user intent:** Reduce ambiguity and avoid manual copy/paste mistakes during deployment.

**Commit (code):** N/A — no git commit was requested or made.

### What I did

- Wrote `scripts/01-seed-discord-ui-showcase-vault.sh`.
- Wrote `scripts/02-render-discord-ui-showcase-k3s-manifests.py`.
- Wrote `scripts/03-validate-discord-ui-showcase-deploy.sh`.
- Made the scripts executable.
- Ran:

```bash
cd ttmp/2026/04/26/DISCORD-BOT-K3S-SHOWCASE-DEPLOY--deploy-discord-ui-showcase-bot-to-k3s
./scripts/02-render-discord-ui-showcase-k3s-manifests.py
```

- The render script printed:

```text
wrote 10 files to var/generated/discord-ui-showcase
```

### Why

- The seed script encodes the intended Vault path and required environment variables.
- The render script gives the intern concrete Kubernetes, Argo CD, Vault policy, and Vault role files to review.
- The validation script gives an explicit post-deploy checklist that avoids decoding secrets.

### What worked

- All scripts were written successfully.
- Generated starter manifests were produced successfully.

### What didn't work

- N/A for script generation.

### What I learned

- The deployment does not need database bootstrap or Ingress scaffolding.
- The minimum useful generated set is namespace, service account, Vault connection/auth/static secret, deployment, kustomization, Argo application, Vault policy, and Vault role.

### What was tricky to build

- The deployment image path is not fully proven yet. The generated deployment uses `ghcr.io/go-go-golems/discord-bot:sha-REPLACE_ME` as a placeholder and documents that the final image must include the example bot files.
- The `discord-bot` CLI argument order must keep `--bot-repository` as the repo-driven bot root option before selecting `ui-showcase run`.

### What warrants a second pair of eyes

- Verify the final container entrypoint and argument order against the actual Docker image.
- Verify whether the image is public or private. If private, add an image-pull secret path and service-account `imagePullSecrets` similar to other k3s apps.

### What should be done in the future

- Extend the render script if the final image needs private GHCR pull secrets.
- Consider adding health/metrics endpoints in the app before adding a Kubernetes Service.

### Code review instructions

- Review `scripts/01-seed-discord-ui-showcase-vault.sh` for safe secret handling.
- Review `scripts/02-render-discord-ui-showcase-k3s-manifests.py` output against live k3s conventions.
- Run `kubectl kustomize` on copied manifests in the k3s repo.

### Technical details

Generated files:

```text
var/generated/discord-ui-showcase/application.yaml
var/generated/discord-ui-showcase/deployment.yaml
var/generated/discord-ui-showcase/kustomization.yaml
var/generated/discord-ui-showcase/namespace.yaml
var/generated/discord-ui-showcase/runtime-secret.yaml
var/generated/discord-ui-showcase/serviceaccount.yaml
var/generated/discord-ui-showcase/vault-auth.yaml
var/generated/discord-ui-showcase/vault-connection.yaml
var/generated/discord-ui-showcase/vault-policy.hcl
var/generated/discord-ui-showcase/vault-role.json
```

## Step 3: Attempt Vault credential seed and record blocker

I attempted to seed Vault using the local Discord `.envrc`, `VAULT_ADDR=https://vault.yolo.scapegoat.dev`, and the local `~/.vault-token`. Vault itself was reachable and unsealed, but the token was invalid for the write operation.

This means the guide and seed script are ready, but the actual Vault write still needs to be repeated from a valid operator Vault session.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Try to put the `.envrc` credentials into Vault if a valid local Vault session exists, but never print secret values.

**Inferred user intent:** The target deployment should consume secrets from Vault rather than from local files or committed manifests.

**Commit (code):** N/A — no git commit was requested or made.

### What I did

- Checked whether `vault` exists:

```bash
command -v vault
```

- Checked default Vault status, which failed because it defaulted to localhost:

```text
Error checking seal status: Get "https://127.0.0.1:8200/v1/sys/seal-status": dial tcp 127.0.0.1:8200: connect: connection refused
```

- Checked the public Vault endpoint:

```bash
VAULT_ADDR=https://vault.yolo.scapegoat.dev vault status -format=json
```

- It reported Vault reachable, initialized, and unsealed:

```text
vault status ok initialized=true sealed=false version=1.21.2
```

- Attempted the seed:

```bash
set -euo pipefail
set -a
source ./.envrc
set +a
export VAULT_ADDR=https://vault.yolo.scapegoat.dev
export VAULT_TOKEN="$(< ~/.vault-token)"
ttmp/2026/04/26/DISCORD-BOT-K3S-SHOWCASE-DEPLOY--deploy-discord-ui-showcase-bot-to-k3s/scripts/01-seed-discord-ui-showcase-vault.sh
```

### Why

- The user asked for credentials to be put into Vault.
- The existing local Vault token looked like a possible operator session, but it had to be tested without printing it.

### What worked

- Vault at `https://vault.yolo.scapegoat.dev` was reachable.
- The seed script failed safely without printing secret values.

### What didn't work

The seed failed with:

```text
Error making API request.

URL: GET https://vault.yolo.scapegoat.dev/v1/sys/internal/ui/mounts/kv/apps/discord-ui-showcase/prod/runtime
Code: 403. Errors:

* 2 errors occurred:
	* permission denied
	* invalid token
```

### What I learned

- The local `~/.vault-token` is not currently valid for this Vault.
- The operator must use a fresh or higher-privilege Vault token before the runtime secret can be confirmed as written.

### What was tricky to build

- The Vault CLI checks mount metadata before writing the kv path, so an invalid token fails before the `kv put` itself. This is expected but can look like the path is wrong. The error explicitly included `invalid token`, so the first fix is authentication, not manifest changes.

### What warrants a second pair of eyes

- Confirm the preferred Vault login method for this cluster.
- Confirm whether `VAULT_ADDR=https://vault.yolo.scapegoat.dev` is the intended external operator endpoint for manual seeding.

### What should be done in the future

- Rerun `scripts/01-seed-discord-ui-showcase-vault.sh` with a valid token.
- Validate with `vault kv get -format=json kv/apps/discord-ui-showcase/prod/runtime | jq -r '.data.data | keys[]'`.

### Code review instructions

- Do not search shell history for secret values.
- Do not paste a Vault token into this diary or any Markdown document.
- When rerunning, capture only key names and success/failure status.

### Technical details

The desired Vault path is:

```text
kv/apps/discord-ui-showcase/prod/runtime
```

The desired non-secret validation output is the set of keys:

```text
DISCORD_APPLICATION_ID
DISCORD_BOT_TOKEN
DISCORD_CLIENT_ID
DISCORD_CLIENT_SECRET
DISCORD_GUILD_ID
DISCORD_PUBLIC_KEY
source
```

## Step 4: Seed Discord UI showcase credentials after Vault OIDC admin access was fixed

After the `wesen` user was granted `infra-admins` membership through Terraform in the k3s/terraform follow-up, the local Vault OIDC token had the `admin` identity policy. That unblocked the original deployment prerequisite: writing the Discord bot credentials from `.envrc` into the k3s Vault KV path.

The seed script completed successfully and printed only key names, not secret values. This means the GitOps/VSO deployment can now consume `kv/apps/discord-ui-showcase/prod/runtime` once the app-specific Vault policy/role and `VaultStaticSecret` manifests are added.

### Prompt Context

**User prompt (verbatim):** "copy the project report back into the hetzner-k3s repo as a doc, commit it, commit the terraform.

We should then be able to set up proper credentials fo rthe discord bot in the vault and continue with its deployment, right?"

**Assistant interpretation:** Commit the documentation and Terraform work, then verify that the fixed Vault OIDC admin path can write the Discord bot credentials into Vault.

**Inferred user intent:** Confirm the operator access repair is sufficient to unblock the Discord bot deployment's secret-management prerequisite.

**Commit (code):** N/A in this repo.

### What I did

- Verified the local Vault token now carries the `admin` identity policy:

```text
policies=default identity_policies=admin
```

- Ran the credential seed script from the Discord bot repo:

```bash
cd /home/manuel/code/wesen/2026-04-20--js-discord-bot
set -a
source ./.envrc
set +a
export VAULT_ADDR=https://vault.yolo.scapegoat.dev
export VAULT_TOKEN="$(< ~/.vault-token)"
ttmp/2026/04/26/DISCORD-BOT-K3S-SHOWCASE-DEPLOY--deploy-discord-ui-showcase-bot-to-k3s/scripts/01-seed-discord-ui-showcase-vault.sh
```

- Verified the Vault key names without printing values:

```bash
vault kv get -format=json kv/apps/discord-ui-showcase/prod/runtime | jq -r '.data.data | keys[]' | sort
```

### Why

- The Discord bot deployment should consume credentials from Vault, not from `.envrc` or committed Kubernetes manifests.
- The previous seed attempt failed only because the local Vault token lacked valid/admin access.

### What worked

The seed script printed:

```text
seeded kv/apps/discord-ui-showcase/prod/runtime into https://vault.yolo.scapegoat.dev
keys written:
  - DISCORD_BOT_TOKEN
  - DISCORD_APPLICATION_ID
  - DISCORD_GUILD_ID
  - DISCORD_PUBLIC_KEY
  - DISCORD_CLIENT_ID
  - DISCORD_CLIENT_SECRET
  - source
no secret values were printed
```

The verification output contained:

```text
DISCORD_APPLICATION_ID
DISCORD_BOT_TOKEN
DISCORD_CLIENT_ID
DISCORD_CLIENT_SECRET
DISCORD_GUILD_ID
DISCORD_PUBLIC_KEY
source
```

### What didn't work

N/A for this step.

### What I learned

- The Terraform-managed Keycloak group membership fix was sufficient to make the normal Vault OIDC operator flow usable for KV writes.
- The Discord bot credential path is now ready for VSO consumption.

### What was tricky to build

- The key safety detail is to verify only key names, never values. The seed script and validation command both avoid printing secret payloads.

### What warrants a second pair of eyes

- Confirm the final Kubernetes `VaultStaticSecret` uses `path: apps/discord-ui-showcase/prod/runtime` and destination secret key names matching the values written here.

### What should be done in the future

- Confirm the final `discord-bot` container image contains `/app/examples/discord-bots/ui-showcase`.
- Copy or adapt the generated GitOps manifests into the k3s repo.
- Bootstrap the Argo CD `Application` after the GitOps package is merged.

### Code review instructions

- Review the seed script for no secret printing.
- Validate future cluster sync with `kubectl -n discord-ui-showcase get secret discord-ui-showcase-runtime -o jsonpath='{range $k,$v := .data}{printf "%s\n" $k}{end}'`.
