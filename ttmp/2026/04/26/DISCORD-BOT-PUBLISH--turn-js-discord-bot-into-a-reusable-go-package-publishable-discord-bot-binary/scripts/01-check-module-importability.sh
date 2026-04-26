#!/usr/bin/env bash
# 01-check-module-importability.sh
# Verifies that the current module can be built and its public API is importable.
#
# Usage:
#   cd /path/to/discord-bot
#   bash ttmp/2026/04/26/DISCORD-BOT-PUBLISH--*/scripts/01-check-module-importability.sh

set -euo pipefail

echo "=== Module Importability Check ==="
echo ""

echo "1. Checking go.mod module path..."
MODULE=$(head -1 go.mod | awk '{print $2}')
echo "   Module: $MODULE"
if echo "$MODULE" | grep -q 'manuel/wesen'; then
    echo "   ⚠️  Still using local development module path"
    echo "   Expected: github.com/go-go-golems/discord-bot"
else
    echo "   ✅ Module path looks canonical"
fi
echo ""

echo "2. Checking for local replace directives..."
if grep -q '^replace.*=>' go.mod; then
    echo "   ⚠️  Found replace directives in go.mod:"
    grep '^replace' go.mod || true
    echo "   These block publishing if they point to local paths."
else
    echo "   ✅ No replace directives found"
fi
echo ""

echo "3. Checking that all packages compile..."
if go build ./... 2>/dev/null; then
    echo "   ✅ go build ./... succeeded"
else
    echo "   ❌ go build ./... failed"
    go build ./... 2>&1 || true
fi
echo ""

echo "4. Checking that tests pass..."
if go test ./internal/... ./pkg/... 2>/dev/null; then
    echo "   ✅ go test ./internal/... ./pkg/... passed"
else
    echo "   ⚠️  Some tests failed (may need credentials for full run)"
fi
echo ""

echo "5. Checking that examples compile..."
if go build ./examples/... 2>/dev/null; then
    echo "   ✅ go build ./examples/... succeeded"
else
    echo "   ❌ go build ./examples/... failed"
    go build ./examples/... 2>&1 || true
fi
echo ""

echo "6. Checking public API surface..."
echo "   pkg/framework/ exports:"
    go doc -all ./pkg/framework/ 2>/dev/null | grep '^func\|^type' | head -20
echo ""
echo "   pkg/botcli/ exports:"
    go doc -all ./pkg/botcli/ 2>/dev/null | grep '^func\|^type' | head -20
echo ""

echo "=== Check complete ==="
