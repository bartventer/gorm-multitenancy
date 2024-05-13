#!/usr/bin/env bash

#-----------------------------------------------------------------------------------------------------------------
# Copyright © 2023 Bart Venter <bartventer@outlook.com>

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#-----------------------------------------------------------------------------------------------------------------
# Maintainer: Bart Venter <https://github.com/bartventer>
#-----------------------------------------------------------------------------------------------------------------
# This script starts local Postgres and pgAdmin instances for testing.
# 
# Usage: ./scripts/localpostgres.sh
#-----------------------------------------------------------------------------------------------------------------

set -euo pipefail

_SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
_POSTGRES_DB="${DB_NAME:-tenants}"

sleep 15

echo ":: Starting Postgres container..."
docker rm -f postgres &>/dev/null || :
docker run -d --name postgres \
    -e POSTGRES_DB="${_POSTGRES_DB}" \
    -e POSTGRES_USER="${DB_USER:-tenants}" \
    -e POSTGRES_PASSWORD="${DB_PASSWORD:-tenants}" \
    -p "${DB_PORT:-5432}":5432 \
    -v ${_SCRIPT_DIR}/testdata:/docker-entrypoint-initdb.d \
    --health-cmd='pg_isready -U $POSTGRES_USER && psql -U $POSTGRES_USER -d $POSTGRES_DB -c "SELECT 1" -q -t' \
    --health-start-period=20s \
    --health-interval=10s \
    --health-timeout=5s \
    --health-retries=5 \
    --restart=always \
    postgres:"${POSTGRES_VERSION:-15.4-alpine}" &>/dev/null
echo "✔️ OK. Run \"docker rm -f postgres\" to clean up the container."
echo

if [[ "${CI:-}" != "true" ]]; then
    echo ":: Starting pgAdmin container..."
    docker rm -f pgadmin &>/dev/null || :
    docker run -d --name pgadmin \
        -e PGADMIN_DEFAULT_EMAIL="${PGADMIN_DEFAULT_EMAIL}" \
        -e PGADMIN_DEFAULT_PASSWORD="${PGADMIN_DEFAULT_PASSWORD}" \
        -p 5050:80 \
        --link postgres:postgres \
        --restart=always \
        dpage/pgadmin4:latest &>/dev/null
    echo "✔️ OK. Run \"docker rm -f pgadmin\" to clean up the container."
    echo
fi