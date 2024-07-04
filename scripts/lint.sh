#!/usr/bin/env bash
set -euo pipefail

echo "================================================================================"
echo "🔍 Starting Linting Process"
echo "================================================================================"

find . -name 'go.mod' -type f \
    -exec echo "--------------------------------------------------------------------------------" \; \
    -exec printf 'Linting in directory: %h\n\n' \; \
    -exec echo "--------------------------------------------------------------------------------" \; \
    -exec echo ":: Initiating linter..." \; \
    -execdir golangci-lint run --fix --verbose ./... \; \
    -exec echo "\n  ✔️ Linting complete.\n" \;

echo "================================================================================"
echo "✅ Linting Process Completed Successfully"
echo "================================================================================"
