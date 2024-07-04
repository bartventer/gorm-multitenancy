#!/usr/bin/env bash
set -euo pipefail

echo "================================================================================"
echo "🔧 Starting Dependency Update Process"
echo "================================================================================"
find . -name 'go.mod' -type f \
    -exec echo "--------------------------------------------------------------------------------" \; \
    -exec printf 'Updating dependencies in: %h\n\n' \; \
    -exec echo "--------------------------------------------------------------------------------" \; \
    -exec echo ":: Running 'go mod tidy' to clean up dependencies..." \; \
    -execdir go mod tidy \; \
    -exec echo "\n  ✔️ Dependencies tidied successfully.\n" \; \
    -exec echo ":: Running 'go get -u ./...' to update dependencies..." \; \
    -execdir go get -u ./... \; \
    -exec echo "\n  ✔️ Dependencies updated successfully.\n" \;

echo "================================================================================"
echo "✅ Dependency Update Process Completed Successfully"
echo "================================================================================"
