#!/usr/bin/env bash
# Starts local Postgres instance for testing.
set -euo pipefail

POSTGRES_VERSION="${POSTGRES_VERSION:-15.4-alpine}"
POSTGRES_DB="${POSTGRES_DB:-"test_tenants"}"
POSTGRES_PORT="${POSTGRES_PORT:-"5432"}"
POSTGRES_HOST="${POSTGRES_HOST:-"localhost"}"
POSTGRES_USER="${POSTGRES_USER:-"postgres"}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-"postgres"}"

echo "
:: Starting Postgres... 

    Configuration:
        * Version: ${POSTGRES_VERSION}
        * Database: ${POSTGRES_DB}
        * Port: ${POSTGRES_PORT}
        * Host: ${POSTGRES_HOST}
        * User: ${POSTGRES_USER}
"
docker rm -f postgres &>/dev/null || :
docker run -d --name postgres \
    -e POSTGRES_DB="${POSTGRES_DB}" \
    -e POSTGRES_USER="${POSTGRES_USER}" \
    -e POSTGRES_PASSWORD="${POSTGRES_PASSWORD}" \
    -p "${POSTGRES_PORT}":5432 \
    postgres:"${POSTGRES_VERSION}" &>/dev/null
echo "OK. Run \"docker rm -f postgres\" to clean up the container."
echo
