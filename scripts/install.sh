#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
INSTALL_DIR="${HOME}/.local/bin"

echo "=== Installing spec ==="
cd "$REPO_ROOT"

mkdir -p "$INSTALL_DIR"
mkdir -p "$REPO_ROOT/data/db/exams"

echo "Building binary with $(go version)..."
rm -f "$INSTALL_DIR/spec"
go build -o "$INSTALL_DIR/spec" ./cmd/spec
chmod +x "$INSTALL_DIR/spec"

# Backward compatibility alias
ln -sf "$INSTALL_DIR/spec" "$INSTALL_DIR/db-quiz"

echo "Verifying installation..."
if command -v spec >/dev/null 2>&1; then
    echo "✓ spec successfully installed to $INSTALL_DIR/spec"
    echo "✓ Path: $(which spec)"
else
    echo "⚠ Installed to $INSTALL_DIR/spec, but $INSTALL_DIR is not in PATH."
fi

echo "=== Installation complete ==="
