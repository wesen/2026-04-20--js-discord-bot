# Changelog

## 2026-04-26

- Initial workspace created.
- Added an intern-oriented deployment design guide for the Discord `ui-showcase` bot.
- Added an investigation diary recording source/k3s evidence, generated scripts, and the Vault seed attempt.
- Added ticket helper scripts for Vault seeding, manifest rendering, and deployment validation.
- Generated starter Argo CD, Kustomize, Vault policy, and Vault role files under `var/generated/discord-ui-showcase/`.
- Attempted to seed `kv/apps/discord-ui-showcase/prod/runtime`; Vault was reachable but the local token failed with `invalid token`, so a valid operator token was required.
- After Terraform-managed `wesen` membership in `infra-admins` was applied in the k3s/terraform follow-up, reran the seed script successfully and verified the expected Vault key names without printing secret values.
- Built and pushed `ghcr.io/go-go-golems/discord-bot:sha-596f442` for `linux/amd64`; the image includes the `discord-bot` binary and `/app/examples/discord-bots`.
- Seeded the private GHCR image pull secret material into Vault at `kv/apps/discord-ui-showcase/prod/image-pull` without printing token values.
- Added and pushed k3s GitOps manifests, Vault policy, and Vault Kubernetes auth role for `discord-ui-showcase` in the Hetzner k3s repository.
- Bootstrapped the Argo CD `discord-ui-showcase` application and verified `Synced Healthy`, VSO runtime/image-pull secret sync, successful rollout, and Discord gateway connection as `llm-bot`.
- Fixed the validation helper script to list Kubernetes Secret keys through `jq` instead of an invalid kubectl JSONPath expression.

## 2026-04-26

Created deployment design guide, diary, helper scripts, generated starter manifests, and recorded Vault seed blocker.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/26/DISCORD-BOT-K3S-SHOWCASE-DEPLOY--deploy-discord-ui-showcase-bot-to-k3s/design-doc/01-discord-ui-showcase-bot-k3s-deployment-design-and-intern-implementation-guide.md — Primary deliverable

