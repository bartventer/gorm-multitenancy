#!/usr/bin/env bash
set -euo pipefail

echo "================================================================================"
echo "🔍 Starting Linting Process"
echo "================================================================================"

find . -name 'go.mod' -type f \
    -printf "--------------------------------------------------------------------------------\n" \
    -printf "Linting in directory: %h\n\n" \
    -printf "--------------------------------------------------------------------------------\n" \
    -printf ":: Initiating linter...\n" \
    -execdir golangci-lint run --fix --verbose ./... \; \
    -printf "\n  ✔️ Linting complete.\n"

echo "================================================================================"
echo "✅ Linting Process Completed Successfully"
echo "================================================================================"
