#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "=== [1/4] Checking Go environment ==="
cd "$REPO_ROOT"
go version

echo "=== [2/4] Running go vet ==="
go vet ./...

echo "=== [3/4] Running go test ==="
go test -v ./...

echo "=== [4/4] Building binary ==="
make build

echo "=== All checks passed successfully! ==="
