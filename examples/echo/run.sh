#!/usr/bin/env bash

# Starts local Postgres instance for the example and runs the example.

set -euo pipefail

# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=tenants
DB_NAME=tenants
DB_PASSWORD=tenants

module_name=$(grep -oP 'module \K(.*)' go.mod)
last_part=$(basename "$module_name")
container_name="postgres-$last_part"

# Find an unused port
while lsof -Pi :$DB_PORT -sTCP:LISTEN -t >/dev/null; do
    DB_PORT=$((DB_PORT + 1))
done

echo ":: Starting local Postgres instance for the $last_part example..."
echo -e "\t- Database: $DB_NAME"
echo -e "\t- User: $DB_USER"
echo -e "\t- Password: $DB_PASSWORD"
echo -e "\t- Host: $DB_HOST"
echo -e "\t- Port: $DB_PORT"
docker rm -f "$container_name" &>/dev/null || :
docker run -d --name "$container_name" \
    -e POSTGRES_DB="$DB_NAME" \
    -e POSTGRES_USER="$DB_USER" \
    -e POSTGRES_PASSWORD="$DB_PASSWORD" \
    -p "$DB_PORT":5432 \
    postgres:"15.4-alpine" &>/dev/null

echo "OK. Run \"docker rm -f $container_name\" to clean up the container."
echo

sleep 10

echo ":: Running the $last_part example..."
DB_HOST=$DB_HOST DB_PORT=$DB_PORT DB_USER=$DB_USER DB_NAME=$DB_NAME DB_PASSWORD=$DB_PASSWORD go run . "$@"
