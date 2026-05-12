#!/usr/bin/env bash
set -euo pipefail

GOOS="${1:-linux}"
GOARCH="${2:-amd64}"
OUTPUT="${3:-pokeface-server}"

echo "==> Building ${GOOS}/${GOARCH} -> ${OUTPUT}..."
CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -o "$OUTPUT" ./cmd/server/

echo "==> Done: ${OUTPUT}"
ls -lh "$OUTPUT"
