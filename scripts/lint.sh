#!/usr/bin/env bash
set -euo pipefail

echo "================================================================================"
echo "🔍 Starting Linting Process"
echo "================================================================================"

gomoddirs=$(find . -name 'go.mod' -type f -exec dirname {} \; | sort)

for dir in $gomoddirs; do
    echo "--------------------------------------------------------------------------------"
    echo "Linting in directory: $dir"
    echo "--------------------------------------------------------------------------------"
    pushd "$dir"
    golangci-lint run --verbose ./...
    popd
done

echo "================================================================================"
echo "✅ Linting Process Completed Successfully"
echo "================================================================================"
