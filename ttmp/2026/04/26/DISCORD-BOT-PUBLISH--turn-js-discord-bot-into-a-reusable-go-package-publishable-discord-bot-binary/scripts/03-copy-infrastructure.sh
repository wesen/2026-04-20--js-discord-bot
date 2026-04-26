#!/usr/bin/env bash
# 03-copy-infrastructure.sh
# Copies infrastructure files from go-template and adapts them for discord-bot.
#
# Usage:
#   cd /path/to/discord-bot
#   bash ttmp/2026/04/26/DISCORD-BOT-PUBLISH--*/scripts/03-copy-infrastructure.sh

set -euo pipefail

TEMPLATE_DIR="${TEMPLATE_DIR:-$HOME/code/wesen/corporate-headquarters/go-template}"

if [ ! -d "$TEMPLATE_DIR" ]; then
    echo "❌ go-template not found at $TEMPLATE_DIR"
    echo "Set TEMPLATE_DIR to the correct path."
    exit 1
fi

echo "=== Copying infrastructure from go-template ==="
echo "  Source: $TEMPLATE_DIR"
echo ""

# Create .github/workflows directory
mkdir -p .github/workflows

# Copy linting config
echo "Copying .golangci.yml..."
cp "$TEMPLATE_DIR/.golangci.yml" .

echo "Copying .golangci-lint-version..."
cp "$TEMPLATE_DIR/.golangci-lint-version" .

# Copy lefthook config
echo "Copying lefthook.yml..."
cp "$TEMPLATE_DIR/lefthook.yml" .

# Copy license
echo "Copying LICENSE..."
cp "$TEMPLATE_DIR/LICENSE" .

# Copy CI workflows
echo "Copying GitHub Actions workflows..."
cp "$TEMPLATE_DIR/.github/workflows/release.yaml" .github/workflows/
cp "$TEMPLATE_DIR/.github/workflows/push.yml" .github/workflows/
cp "$TEMPLATE_DIR/.github/workflows/lint.yml" .github/workflows/
cp "$TEMPLATE_DIR/.github/workflows/codeql-analysis.yml" .github/workflows/
cp "$TEMPLATE_DIR/.github/workflows/secret-scanning.yml" .github/workflows/
cp "$TEMPLATE_DIR/.github/workflows/dependency-scanning.yml" .github/workflows/

echo ""
echo "✅ Infrastructure files copied."
echo ""
echo "⚠️  You still need to manually create:"
echo "  - Makefile (adapt from go-template, replace XXX with discord-bot)"
echo "  - .goreleaser.yaml (adapt from go-template)"
echo "  - AGENT.md (adapt from go-template)"
echo ""
echo "See design-doc/05-infrastructure-cicd.md for the exact contents."
