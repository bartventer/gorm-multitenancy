#!/usr/bin/env bash

#-----------------------------------------------------------------------------------------------------------------
# Copyright (c) Bart Venter.
# Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance
# with the License. You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
#-----------------------------------------------------------------------------------------------------------------
# Maintainer: Bart Venter <https://github.com/bartventer>
#-----------------------------------------------------------------------------------------------------------------
# This script checks if all services are healthy. It accepts a list of services to skip as an argument.
# 
# Usage: ./scripts/check_services.sh "service1 service2 service3"
#-----------------------------------------------------------------------------------------------------------------

set -euo pipefail

_SERVICES_TO_SKIP=${1:-""} # Use the first argument as the list of services to skip, or an empty string if no argument is provided
_MAX_ATTEMPT=30
_SLEEP_BETWEEN_ATTEMPTS=2
_ALL_SERVICES_HEALTHY=true

# Get the list of running Docker containers
_SERVICES=$(docker ps --format '{{.Names}}')

if (( ${#_SERVICES} == 0 )); then
    echo "ERROR: No services are running."
    exit 1
fi

# Function to check the health of a single service
check_service_health() {
    local service_name=$1
    echo "INFO: Checking service \"${service_name}\"…"
    local health_start_period
    health_start_period=$(docker inspect "${service_name}" | jq -r '(.[] | .Config.Healthcheck.StartPeriod // 0) / 1e9')
    if (( health_start_period > 0 )); then
        echo "INFO: Waiting for ${health_start_period} seconds before starting health checks…"
        sleep ${health_start_period}
    fi

    attempt_counter=1
    until
        docker inspect "${service_name}" | jq -r '.[] | [.State.Status, ":", .State.Health.Status] | add' |
            grep -q -e '^running:$' -e '^running:healthy$'
    do
        sleep ${_SLEEP_BETWEEN_ATTEMPTS}
        if [[ attempt_counter -ge ${_MAX_ATTEMPT} ]]; then
            echo "WARN: Service \"${service_name}\" is still not healthy after #${attempt_counter} attempts."
            _ALL_SERVICES_HEALTHY=false
            break
        fi
        echo "INFO: Service \"${service_name}\" is not (yet) healthy (attempt #${attempt_counter}): Will try again…"
        ((attempt_counter++))
    done
    if [[ $_ALL_SERVICES_HEALTHY == true ]]; then
        echo "INFO: ✔️ Service \"${service_name}\" is healthy."
    fi
}

# Run the health checks in parallel
for service_name in $_SERVICES; do
    if [[ $_SERVICES_TO_SKIP =~ (^|[[:space:]])"$service_name"($|[[:space:]]) ]]; then
        echo "INFO: Skipping running check for service \"${service_name}\"."
        continue
    fi
    check_service_health "${service_name}" &
done

# Wait for all background processes to complete
wait

if [[ $_ALL_SERVICES_HEALTHY != true ]]; then
    echo "ERROR: Not all services are healthy."
    exit 1
else
    echo "INFO: ✅ All services are healthy."
fi