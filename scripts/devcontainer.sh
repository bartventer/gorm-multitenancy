#!/usr/bin/env bash

set -euo pipefail

echo "🚀 Building devcontainer..."
devcontainer build \
    --log-level debug \
    --workspace-folder .devcontainer/build \
    --image-name ghcr.io/bartventer/gorm-multitenancy/devcontainer:latest \
    --platform linux/amd64 \
    --push

echo "🎉 OK. Successfully built devcontainer."
