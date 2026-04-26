#!/usr/bin/env bash
# 02-rename-module-path.sh
# Renames the Go module path from the local development name to the canonical go-go-golems path.
#
# Usage:
#   cd /path/to/js-discord-bot
#   bash ttmp/2026/04/26/DISCORD-BOT-PUBLISH--*/scripts/02-rename-module-path.sh [--dry-run]

set -euo pipefail

OLD_MODULE="github.com/manuel/wesen/2026-04-20--js-discord-bot"
NEW_MODULE="github.com/go-go-golems/discord-bot"

DRY_RUN=false
if [[ "${1:-}" == "--dry-run" ]]; then
    DRY_RUN=true
    echo "=== DRY RUN ==="
fi

echo "=== Renaming module path ==="
echo "  From: $OLD_MODULE"
echo "  To:   $NEW_MODULE"
echo ""

# Find all Go files (excluding .git, ttmp, testdata)
FILES=$(find . -name '*.go' -not -path './.git/*' -not -path './ttmp/*' -not -path './testdata/*')

COUNT=$(echo "$FILES" | xargs grep -l "$OLD_MODULE" 2>/dev/null | wc -l)
echo "Found $COUNT files containing the old module path."
echo ""

if [ "$DRY_RUN" = true ]; then
    echo "Files that would be changed:"
    echo "$FILES" | xargs grep -l "$OLD_MODULE" 2>/dev/null || true
    echo ""
    echo "To apply, run without --dry-run."
    exit 0
fi

# Replace in go.mod
echo "Updating go.mod..."
sed -i "s|$OLD_MODULE|$NEW_MODULE|g" go.mod

# Replace in all Go files
echo "Updating Go source files..."
echo "$FILES" | xargs sed -i "s|$OLD_MODULE|$NEW_MODULE|g"

echo ""
echo "=== Verifying ==="
REMAINING=$(echo "$FILES" | xargs grep -l "$OLD_MODULE" 2>/dev/null | wc -l)
if [ "$REMAINING" -eq 0 ]; then
    echo "✅ All references updated."
else
    echo "⚠️  $REMAINING files still contain the old path:"
    echo "$FILES" | xargs grep -l "$OLD_MODULE" 2>/dev/null
fi

echo ""
echo "Running go mod tidy..."
go mod tidy

echo ""
echo "Running go build ./..."
if go build ./...; then
    echo "✅ Build succeeded after rename."
else
    echo "❌ Build failed. Check the errors above."
    exit 1
fi

echo ""
echo "=== Rename complete ==="
echo "Next: git add -A && git commit -m 'chore: rename module to $NEW_MODULE'"
