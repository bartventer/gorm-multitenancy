#!/usr/bin/env bash
set -euo pipefail

echo "================================================================================"
echo "🚀 Benchmarking Start"
echo "================================================================================"

find . -name 'go.mod' -type f \
    -exec echo "--------------------------------------------------------------------------------" \; \
    -exec printf 'Benchmarking in directory: %h\n\n' \; \
    -exec echo "--------------------------------------------------------------------------------" \; \
    -exec echo ":: Executing benchmarks..." \; \
    -execdir go test -v -run=^$ -bench=. -benchmem -tags=gorm_multitenancy_benchmarks ./... \; \
    -exec echo "\n  ✔️ Benchmarking complete.\n" \;

echo "================================================================================"
echo "✅ Benchmarking Completed Successfully"
echo "================================================================================"
