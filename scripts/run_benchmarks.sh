#!/usr/bin/env bash

#-----------------------------------------------------------------------------------------------------------------
# Copyright (c) Bart Venter.
# Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance
# with the License. You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
#-----------------------------------------------------------------------------------------------------------------
# Maintainer: Bart Venter <https://github.com/bartventer>
#-----------------------------------------------------------------------------------------------------------------
# This script runs go benchmarks for the provided functions in the provided package package.
# The following flags are required:
#
#       -package package1
#           The package to run the benchmarks in.
#
#       -benchfuncs "BenchmarkFunc1,BenchmarkFunc2"
#           A comma-separated list of benchmark functions to run.
#
#       -outputdir /path/to/outputdir
#           The path to the directory where the benchmark results will be saved.
#
# Usage: ./run_benchmarks.sh -package [target_pkg] -benchfuncs [bench_funcs] -outputdir [outputdir]
#
# Example: ./run_benchmarks.sh -package mypkg -benchfuncs "BenchmarkFunc1,BenchmarkFunc2" -outputdir ./output
#-----------------------------------------------------------------------------------------------------------------

set -euo pipefail

# Default values
_PKG=""
_BENCH_FUNCS=()
_OUTPUTDIR=""

# Parse flags
while (("$#")); do
    case "$1" in
    -package)
        _PKG="$2"
        shift 2
        ;;
    -benchfuncs)
        IFS=',' read -r -a _BENCH_FUNCS <<<"$2"
        shift 2
        ;;
    -outputdir)
        _OUTPUTDIR="$2"
        shift 2
        ;;
    *)
        echo "Error: Invalid flag $1"
        exit 1
        ;;
    esac
done

# Check if all required flags were provided
if [[ -z $_PKG || ${#_BENCH_FUNCS[@]} -eq 0 || -z $_OUTPUTDIR ]]; then
    echo "
Error: Missing required flags.

Usage: ./run_benchmarks.sh -package [target_pkg] -benchfuncs [bench_funcs] -outputdir [outputdir]
"
    exit 1
fi

echo "ℹ️ Running benchmarks for \"${_BENCH_FUNCS[*]}\" in \"$_PKG\"..."

mkdir -p "$_OUTPUTDIR"

for bench in "${_BENCH_FUNCS[@]}"; do
    go test -bench=^"$bench"$ -run=^$ -benchmem -benchtime=2s "$_PKG" |
        tee >(grep -v "^PASS\|^ok" >"$_OUTPUTDIR/$bench.txt")
done

echo "Benchmarks saved to \"$_OUTPUTDIR\"."
echo "✔️ OK."

format_output() {
    local input_file="$1"
    local output_file="$2"
    local date=$(date +"%Y-%m-%d")

    echo -e "#### Benchmarks\n" >"$output_file"
    while IFS= read -r line; do
        if [[ $line == goos:* || $line == goarch:* || $line == pkg:* || $line == cpu:* ]]; then
            echo "- $line" >>"$output_file"
        elif [[ $line == Benchmark* ]]; then
            echo -e "- date: $date\n\n| Benchmark | ns/op | B/op | allocs/op |\n|-----------|-------|------|-----------|" >>"$output_file"
            break
        fi
    done <"$input_file"

    while IFS= read -r line; do
        if [[ $line == Benchmark* ]]; then
            local benchmark=$(echo "$line" | awk '{print $1}')
            local ns_op=$(echo "$line" | awk '{print $3}')
            local b_op=$(echo "$line" | awk '{print $5}')
            local allocs_op=$(echo "$line" | awk '{print $7}')
            echo "| $benchmark | $ns_op | $b_op | $allocs_op |" >>"$output_file"
        fi
    done <"$input_file"
}

for bench in "${_BENCH_FUNCS[@]}"; do
    format_output "$_OUTPUTDIR/$bench.txt" "$_OUTPUTDIR/$bench.md"
done