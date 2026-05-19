#!/bin/bash
set -e

cd "$(dirname "$0")/../.."
PROJECT_ROOT="$PWD"

mkdir -p tests/vscode/output

echo "==> Building VS Code test image..."
docker build -f tests/vscode/Dockerfile -t temporal-lsp-vscode-test .

echo "==> Running VS Code screenshot tests..."
docker run --rm \
  -v "$PROJECT_ROOT/tests/vscode/output:/output" \
  temporal-lsp-vscode-test

echo ""
echo "==> Done! Screenshots:"
ls -la tests/vscode/output/*.png 2>/dev/null
echo ""
echo "Videos (if captured):"
ls -la tests/vscode/output/**/*.webm 2>/dev/null || true
