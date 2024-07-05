#!/usr/bin/env bash
set -euo pipefail

echo "================================================================================"
echo "🔧 Starting Dependency Update Process"
echo "================================================================================"
find . -name 'go.mod' -type f \
    -printf "--------------------------------------------------------------------------------\n" \
    -printf "Updating dependencies in: %h\n\n" \
    -printf "--------------------------------------------------------------------------------\n" \
    -printf ":: Running 'go mod tidy' to clean up dependencies...\n" \
    -execdir go mod tidy \; \
    -printf "\n  ✔️ Dependencies tidied successfully.\n" \
    -printf ":: Running 'go get -u ./...' to update dependencies...\n" \
    -execdir go get -u ./... \; \
    -printf "\n  ✔️ Dependencies updated successfully.\n"

echo "================================================================================"
echo "✅ Dependency Update Process Completed Successfully"
echo "================================================================================"
