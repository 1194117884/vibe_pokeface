#!/usr/bin/env bash
set -euo pipefail

# Harness verification gate: lint → typecheck → test → build
# Stops on first failure.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "=== VERIFY: Server (Go) ==="
cd "$PROJECT_ROOT/server"

echo "→ go vet ./..."
go vet ./...

echo "→ go test ./..."
go test ./...

echo "=== VERIFY: Frontend (TypeScript) ==="
cd "$PROJECT_ROOT/frontend"

echo "→ npm run lint"
npm run lint

echo "→ tsc --noEmit"
npx tsc --noEmit

echo "→ npm test"
npm test

echo ""
echo "✓ All checks passed"
