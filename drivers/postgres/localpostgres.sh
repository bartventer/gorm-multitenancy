#!/usr/bin/env bash

# Starts local Postgres instance for testing.

set -euo pipefail

echo ":: Starting Postgres container..."
docker rm -f postgres &>/dev/null || :
docker run -d --name postgres \
    -e POSTGRES_DB="${POSTGRES_DB:-test_tenants}" \
    -e POSTGRES_USER="${POSTGRES_USER:-postgres}" \
    -e POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-postgres}" \
    -p "${POSTGRES_PORT:-5432}":5432 \
    postgres:"${POSTGRES_VERSION:-15.4-alpine}" &>/dev/null
echo "OK. Run \"docker rm -f postgres\" to clean up the container."
echo