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
# This script updates all go mod files across the gorm-multitenancy repository.
# The following flags are optional:
#
#       -tags tag1,tag2
#           A comma-separated list of build tags to consider satisfied during the build.
#
# Usage: ./update_deps.sh [-tags tag1,tag2]
#
# Example: ./update_deps.sh -tags mytag1,mytag2
#-----------------------------------------------------------------------------------------------------------------

set -euo pipefail

# Default values
_TAGS=""

# Parse flags
while (("$#")); do
    case "$1" in
    -tags)
        _TAGS="$2"
        shift 2
        ;;
    *)
        echo "Error: Invalid flag $1"
        exit 1
        ;;
    esac
done

echo "ℹ️ Starting go mod files update..."

find . -name 'go.mod' -not -path './vendor/*' \
    -printf '\n\n:: Updating go modules for %p...' \
    -execdir go get -u -v -tags ${_TAGS} \; \
    -printf '\n  ✔️ OK. Updated.' \
    -printf '\n\n:: Tidying go modules for %p...' \
    -execdir go mod tidy -v \; \
    -printf '\n  ✔️ OK. Tidied.'

printf "\n\n✅ OK. Go mod files update complete."