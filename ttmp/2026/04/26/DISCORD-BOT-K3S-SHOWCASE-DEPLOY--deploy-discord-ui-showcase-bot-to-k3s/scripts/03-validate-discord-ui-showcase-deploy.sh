#!/usr/bin/env bash

set -euo pipefail

app="${DISCORD_UI_SHOWCASE_APP:-discord-ui-showcase}"
namespace="${DISCORD_UI_SHOWCASE_NAMESPACE:-discord-ui-showcase}"
secret_name="${DISCORD_UI_SHOWCASE_SECRET_NAME:-discord-ui-showcase-runtime}"

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

require_cmd kubectl
require_env KUBECONFIG

echo "Argo application status:"
kubectl -n argocd get application "${app}" -o jsonpath='{.status.sync.status} {.status.health.status}{"\n"}'

echo "Vault-rendered Kubernetes Secret keys:"
kubectl -n "${namespace}" get secret "${secret_name}" -o jsonpath='{range $k,$v := .data}{printf "  - %s\n" $k}{end}'

echo "Deployment rollout:"
kubectl -n "${namespace}" rollout status "deployment/${app}" --timeout=180s

echo "Recent pod summary:"
kubectl -n "${namespace}" get pods -l "app.kubernetes.io/name=${app}" -o wide

echo "Recent logs (secrets should already be redacted by the app):"
kubectl -n "${namespace}" logs -l "app.kubernetes.io/name=${app}" --tail=80 --prefix=true
