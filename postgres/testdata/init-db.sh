#!/usr/bin/env bash

#-----------------------------------------------------------------------------------------------------------------
# Copyright (c) Bart Venter.
# Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance
# with the License. You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
#-----------------------------------------------------------------------------------------------------------------
# Maintainer: Bart Venter <https://github.com/bartventer>
#-----------------------------------------------------------------------------------------------------------------
# This script creates multiple tenant databases. It accepts the Postgres username as an environment variable.
# 
# Usage: POSTGRES_USER=myuser ./init-db.sh
#-----------------------------------------------------------------------------------------------------------------

set -euo pipefail

POSTGRES_USER=${POSTGRES_USER:-tenants}

for i in {1..2}; do
  psql -U "$POSTGRES_USER" -c "CREATE DATABASE tenants$i OWNER $POSTGRES_USER;"
done