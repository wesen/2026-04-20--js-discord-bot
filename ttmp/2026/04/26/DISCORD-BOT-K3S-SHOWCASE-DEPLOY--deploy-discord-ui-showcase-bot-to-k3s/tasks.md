# Tasks

## TODO

- [x] Create the docmgr ticket workspace.
- [x] Add the primary design and implementation guide.
- [x] Add the chronological investigation diary.
- [x] Inspect the Discord bot runtime, selected `ui-showcase` bot, and k3s deployment patterns.
- [x] Add helper scripts under the ticket `scripts/` directory.
- [x] Generate starter GitOps/Vault manifests under `var/generated/discord-ui-showcase/`.
- [x] Attempt Vault credential seeding without printing secret values.
- [x] Rerun Vault credential seeding with a valid operator token after fixing the `wesen` Vault OIDC admin membership.
- [x] Copy reviewed manifests into `/home/manuel/code/wesen/2026-03-27--hetzner-k3s` after confirming the final image tag.
- [x] Validate and bootstrap the Argo CD `Application` after the GitOps PR is merged.
- [x] Build and push `ghcr.io/go-go-golems/discord-bot:sha-596f442` containing the `ui-showcase` example bot files.
- [x] Seed the GHCR image pull secret into Vault at `kv/apps/discord-ui-showcase/prod/image-pull` without printing token values.
- [x] Apply the Argo CD `discord-ui-showcase` application and verify `Synced Healthy`, VSO secret sync, rollout, and Discord gateway connection.
