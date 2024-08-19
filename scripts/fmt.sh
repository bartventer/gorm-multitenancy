#!/usr/bin/env bash
set -euo pipefail

gomoddirs=$(find . -name 'go.mod' -type f -exec dirname {} \; | sort)

echo "================================================================================"
echo "🔍 Starting Formatting Process"
echo "================================================================================"

for dir in $gomoddirs; do
    echo "--------------------------------------------------------------------------------"
    echo "Formatting in directory: $dir"
    echo "--------------------------------------------------------------------------------"
    pushd "$dir"
    golangci-lint run --fix --verbose ./...
    popd
done

echo "================================================================================"
echo "✅ Formatting Process Completed Successfully"
echo "================================================================================"
