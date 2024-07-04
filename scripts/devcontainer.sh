#!/usr/bin/env bash
set -euo pipefail

echo "================================================================================"
echo "🚀 Starting Devcontainer Build Process"
echo "================================================================================"
devcontainer build \
    --log-level debug \
    --workspace-folder .devcontainer/build \
    --image-name ghcr.io/bartventer/gorm-multitenancy/devcontainer:latest \
    --platform linux/amd64 \
    --push

echo "================================================================================"
echo "🎉 Devcontainer Build Process Completed Successfully"
echo "================================================================================"
