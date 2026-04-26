#!/usr/bin/env python3
"""Render starter GitOps manifests for the discord-ui-showcase k3s deployment.

The script writes a reviewable Kustomize package under an output directory. It is
safe to run repeatedly in the ticket workspace and then copy into the Hetzner k3s
repo when the operator is ready.
"""

from __future__ import annotations

import argparse
from pathlib import Path

APP = "discord-ui-showcase"
NAMESPACE = APP
SECRET_NAME = f"{APP}-runtime"
IMAGE = "ghcr.io/go-go-golems/discord-bot:sha-REPLACE_ME"

FILES = {
    "namespace.yaml": f"""apiVersion: v1
kind: Namespace
metadata:
  name: {NAMESPACE}
""",
    "serviceaccount.yaml": f"""apiVersion: v1
kind: ServiceAccount
metadata:
  name: {APP}
""",
    "vault-connection.yaml": """apiVersion: secrets.hashicorp.com/v1beta1
kind: VaultConnection
metadata:
  name: vault
  annotations:
    argocd.argoproj.io/sync-wave: "-2"
spec:
  address: http://vault.vault.svc.cluster.local:8200
  skipTLSVerify: true
""",
    "vault-auth.yaml": f"""apiVersion: secrets.hashicorp.com/v1beta1
kind: VaultAuth
metadata:
  name: {APP}
  annotations:
    argocd.argoproj.io/sync-wave: "-2"
spec:
  vaultConnectionRef: vault
  method: kubernetes
  mount: kubernetes
  kubernetes:
    role: {APP}
    serviceAccount: {APP}
""",
    "runtime-secret.yaml": f"""apiVersion: secrets.hashicorp.com/v1beta1
kind: VaultStaticSecret
metadata:
  name: {SECRET_NAME}
  annotations:
    argocd.argoproj.io/sync-wave: "-1"
spec:
  vaultAuthRef: {APP}
  mount: kv
  type: kv-v2
  path: apps/{APP}/prod/runtime
  refreshAfter: 30s
  destination:
    name: {SECRET_NAME}
    create: true
    overwrite: true
""",
    "deployment.yaml": f"""apiVersion: apps/v1
kind: Deployment
metadata:
  name: {APP}
  annotations:
    argocd.argoproj.io/sync-wave: "2"
  labels:
    app.kubernetes.io/name: {APP}
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: {APP}
  template:
    metadata:
      labels:
        app.kubernetes.io/name: {APP}
    spec:
      enableServiceLinks: false
      serviceAccountName: {APP}
      containers:
        - name: {APP}
          image: {IMAGE}
          imagePullPolicy: IfNotPresent
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
                  name: {SECRET_NAME}
                  key: DISCORD_BOT_TOKEN
            - name: DISCORD_APPLICATION_ID
              valueFrom:
                secretKeyRef:
                  name: {SECRET_NAME}
                  key: DISCORD_APPLICATION_ID
            - name: DISCORD_GUILD_ID
              valueFrom:
                secretKeyRef:
                  name: {SECRET_NAME}
                  key: DISCORD_GUILD_ID
                  optional: true
            - name: DISCORD_PUBLIC_KEY
              valueFrom:
                secretKeyRef:
                  name: {SECRET_NAME}
                  key: DISCORD_PUBLIC_KEY
                  optional: true
            - name: DISCORD_CLIENT_ID
              valueFrom:
                secretKeyRef:
                  name: {SECRET_NAME}
                  key: DISCORD_CLIENT_ID
                  optional: true
            - name: DISCORD_CLIENT_SECRET
              valueFrom:
                secretKeyRef:
                  name: {SECRET_NAME}
                  key: DISCORD_CLIENT_SECRET
                  optional: true
          resources:
            requests:
              cpu: 50m
              memory: 128Mi
            limits:
              memory: 512Mi
""",
    "kustomization.yaml": f"""apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: {NAMESPACE}
resources:
  - namespace.yaml
  - serviceaccount.yaml
  - vault-connection.yaml
  - vault-auth.yaml
  - runtime-secret.yaml
  - deployment.yaml
""",
    "application.yaml": f"""apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: {APP}
  namespace: argocd
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: default
  destination:
    server: https://kubernetes.default.svc
    namespace: {NAMESPACE}
  source:
    repoURL: https://github.com/wesen/2026-03-27--hetzner-k3s.git
    targetRevision: main
    path: gitops/kustomize/{APP}
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
      - ServerSideApply=true
""",
    "vault-policy.hcl": f"""path \"kv/data/apps/{APP}/prod/*\" {{
  capabilities = [\"read\"]
}}

path \"kv/metadata/apps/{APP}/prod/*\" {{
  capabilities = [\"read\", \"list\"]
}}
""",
    "vault-role.json": f"""{{
  \"bound_service_account_names\": [\"{APP}\"],
  \"bound_service_account_namespaces\": [\"{NAMESPACE}\"],
  \"policies\": [\"{APP}\"],
  \"ttl\": \"1h\",
  \"max_ttl\": \"4h\"
}}
""",
}


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--out", default="var/generated/discord-ui-showcase", help="output directory")
    args = parser.parse_args()

    out = Path(args.out)
    out.mkdir(parents=True, exist_ok=True)
    for name, content in FILES.items():
        (out / name).write_text(content, encoding="utf-8")
    print(f"wrote {len(FILES)} files to {out}")


if __name__ == "__main__":
    main()
