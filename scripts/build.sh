#!/usr/bin/env bash
set -euo pipefail

echo "==============================================================================="
echo "🛠️  Build process started"
echo "==============================================================================="
find . -name 'go.mod' -type f \
    -printf "-------------------------------------------------------------------------------\n" \
    -printf "Found go.mod in: %h\n\n" \
    -printf "-------------------------------------------------------------------------------\n" \
    -printf ":: Starting build in directory...\n" \
    -execdir go build -v -o /dev/null ./... \; \
    -printf "\n  ✔️ Build successful.\n"

echo "==============================================================================="
echo "✅ Build process completed"
echo "==============================================================================="