#!/usr/bin/env bash
set -euo pipefail

echo "================================================================================"
echo "🔧 Starting Dependency Update Process"
echo "================================================================================"

gomoddirs=$(find . -name 'go.mod' -type f -exec dirname {} \; | sort)
for dir in $gomoddirs; do
    echo "--------------------------------------------------------------------------------"
    echo "Updating dependencies in: $dir"
    echo "--------------------------------------------------------------------------------"
    echo ":: Running 'go mod tidy' to clean up dependencies..."
    (cd "$dir" && go mod tidy)
    echo "  ✔️ Dependencies tidied successfully."
    echo ":: Running 'go get -u ./...' to update dependencies..."
    (cd "$dir" && go get -u ./...)
    echo "  ✔️ Dependencies updated successfully."
done

echo "================================================================================"
echo "✅ Dependency Update Process Completed Successfully"
echo "================================================================================"
