#!/usr/bin/env bash

#-----------------------------------------------------------------------------------------------------------------
# Copyright (c) Bart Venter.
# Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance
# with the License. You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
#-----------------------------------------------------------------------------------------------------------------
# Maintainer: Bart Venter <https://github.com/bartventer>
#-----------------------------------------------------------------------------------------------------------------
# This script starts local dependencies for gorm-multitenancy tests.
#-----------------------------------------------------------------------------------------------------------------

set -euo pipefail

echo ":: Starting local dependencies setup..."
echo
# shellcheck disable=SC1091
# shellcheck source=drivers/postgres/localpostgres.sh
. ./drivers/postgres/localpostgres.sh

echo "✔️ OK. Local dependencies setup complete."
echo
