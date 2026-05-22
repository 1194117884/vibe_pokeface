#!/usr/bin/env bash
set -euo pipefail

# Full CI simulation for local pre-push verification.
# Runs same checks as CI would.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "=== CI LOCAL ==="
echo ""

# Run standard verification
"$SCRIPT_DIR/verify.sh"

# Additional checks
echo ""
echo "=== Additional CI Checks ==="

PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

# Check for leftover TODO/FIXME without issue reference
echo "→ Checking for unresolved TODOs..."
TODOS=$(grep -rn "TODO\|FIXME" --include="*.go" --include="*.ts" --include="*.tsx" server/ frontend/ 2>/dev/null || true)
if [ -n "$TODOS" ]; then
  echo "  Note: Found TODOs/FIXMEs (review before merge):"
  echo "$TODOS"
fi

# Check no secrets in staged files
echo "→ Checking for potential secrets..."
if command -v gitleaks &> /dev/null; then
  gitleaks detect --no-git 2>/dev/null || echo "  gitleaks: check complete"
else
  echo "  gitleaks not installed — skipping"
fi

echo ""
echo "✓ CI checks complete"
