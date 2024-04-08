#!/usr/bin/env bash

# Starts local dependencies for gorm-multitenancy tests.

set -euo pipefail

echo ":: Starting local dependencies setup..."
# shellcheck disable=SC1091
# shellcheck source=drivers/postgres/localpostgres.sh
. ./drivers/postgres/localpostgres.sh

sleep 10

echo "OK. Local dependencies setup complete."
