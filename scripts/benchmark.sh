#!/usr/bin/env bash
set -euo pipefail

echo "================================================================================"
echo "🚀 Benchmarking Start"
echo "================================================================================"

find . -name 'go.mod' -type f \
    -printf "--------------------------------------------------------------------------------\n" \
    -printf "Benchmarking in directory: %h\n\n" \
    -printf "--------------------------------------------------------------------------------\n" \
    -printf ":: Executing benchmarks...\n" \
    -execdir go test -v -run=^$ -bench=. -benchmem -tags=gorm_multitenancy_benchmarks ./... \; \
    -printf "\n  ✔️ Benchmarking complete.\n"

echo "================================================================================"
echo "✅ Benchmarking Completed Successfully"
echo "================================================================================"
