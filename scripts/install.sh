#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
INSTALL_DIR="${HOME}/.local/bin"

echo "=== Installing db-quiz ==="
cd "$REPO_ROOT"

mkdir -p "$INSTALL_DIR"
mkdir -p "$REPO_ROOT/data/exams"

echo "Building binary with $(go version)..."
rm -f "$INSTALL_DIR/db-quiz"
go build -o "$INSTALL_DIR/db-quiz" ./cmd/db-quiz
chmod +x "$INSTALL_DIR/db-quiz"

echo "Verifying installation..."
if command -v db-quiz >/dev/null 2>&1; then
    echo "✓ db-quiz successfully installed to $INSTALL_DIR/db-quiz"
    echo "✓ Path: $(which db-quiz)"
else
    echo "⚠ Installed to $INSTALL_DIR/db-quiz, but $INSTALL_DIR is not in PATH."
    echo "  Please add 'export PATH=\"\$HOME/.local/bin:\$PATH\"' to your ~/.zshrc or ~/.bashrc."
fi

echo "=== Installation complete ==="
