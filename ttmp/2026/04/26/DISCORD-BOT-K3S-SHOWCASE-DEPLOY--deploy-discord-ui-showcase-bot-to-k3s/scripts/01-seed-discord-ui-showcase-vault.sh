#!/usr/bin/env bash

set -euo pipefail

# Seed the Discord UI showcase bot credentials from this repository's .envrc into Vault.
# This script intentionally prints only paths and key names, never secret values.
#
# Expected operator environment:
#   VAULT_ADDR and VAULT_TOKEN are already exported, usually by sourcing the k3s repo .envrc.
#   DISCORD_* variables are exported, usually by sourcing the discord-bot repo .envrc.
#
# Example:
#   set -a
#   source /home/manuel/code/wesen/2026-03-27--hetzner-k3s/.envrc
#   source /home/manuel/code/wesen/2026-04-20--js-discord-bot/.envrc
#   set +a
#   ./scripts/01-seed-discord-ui-showcase-vault.sh

require_cmd() {
  local name="$1"
  command -v "$name" >/dev/null 2>&1 || {
    echo "required command not found: $name" >&2
    exit 1
  }
}

require_env() {
  local name="$1"
  [[ -n "${!name:-}" ]] || {
    echo "missing required environment variable: $name" >&2
    exit 1
  }
}

optional_kv_arg() {
  local key="$1"
  local env_name="$2"
  if [[ -n "${!env_name:-}" ]]; then
    printf '%s=%s\n' "$key" "${!env_name}"
  fi
}

require_cmd vault
require_env VAULT_ADDR
require_env VAULT_TOKEN
require_env DISCORD_BOT_TOKEN
require_env DISCORD_APPLICATION_ID

kv_mount_path="${VAULT_KV_MOUNT_PATH:-kv}"
secret_path="${DISCORD_UI_SHOWCASE_RUNTIME_SECRET_PATH:-apps/discord-ui-showcase/prod/runtime}"

args_file="$(mktemp)"
trap 'rm -f "$args_file"' EXIT

optional_kv_arg DISCORD_BOT_TOKEN DISCORD_BOT_TOKEN >>"$args_file"
optional_kv_arg DISCORD_APPLICATION_ID DISCORD_APPLICATION_ID >>"$args_file"
optional_kv_arg DISCORD_GUILD_ID DISCORD_GUILD_ID >>"$args_file"
optional_kv_arg DISCORD_PUBLIC_KEY DISCORD_PUBLIC_KEY >>"$args_file"
optional_kv_arg DISCORD_CLIENT_ID DISCORD_CLIENT_ID >>"$args_file"
optional_kv_arg DISCORD_CLIENT_SECRET DISCORD_CLIENT_SECRET >>"$args_file"
printf '%s=%s\n' source "01-seed-discord-ui-showcase-vault.sh" >>"$args_file"

# shellcheck disable=SC2046
vault kv put "${kv_mount_path}/${secret_path}" $(<"$args_file") >/dev/null

echo "seeded ${kv_mount_path}/${secret_path} into ${VAULT_ADDR}"
echo "keys written:"
cut -d= -f1 "$args_file" | sed 's/^/  - /'
echo "no secret values were printed"
