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
# This script runs go benchmarks for the provided functions in the provided package package.
# The following flags are required:
#
#       -package package1
#           The package to run the benchmarks in.
#
#       -benchfunc "BenchmarkFunc1"
#           The benchmark function to run.
#
#       -outputdir /path/to/outputdir
#           The directory where the benchmark results will be saved.
#
#       -template /path/to/template
#           The path to the template file used to format the benchmark results.
#
# Usage: ./benchmarks.sh -package [target_pkg] -benchfunc [bench_func] -outputdir [outputdir]
#
# Example: ./benchmarks.sh -package mypkg -benchfunc "BenchmarkFunc1" -outputdir ./outputdir
#-----------------------------------------------------------------------------------------------------------------

set -euo pipefail

# Default values
_PKG=""
_BENCH_FUNC=""
_OUTPUTDIR=""
_TEMPLATE="benchmark_template.md"

# usage prints the usage of the script
usage() {
    echo "
Usage: ./run_benchmarks.sh -package [target_pkg] -benchfunc [bench_func] -outputdir [outputdir]

Flags:
    -package
        The package to run the benchmarks in.

    -benchfunc
        The benchmark function to run.

    -outputdir
        The directory where the benchmark results will be saved.

    -template
        The path to the template file used to format the benchmark results.
        Default: benchmark_template.m4
"
}

# Parse flags
while (("$#")); do
    case "$1" in
    -package)
        _PKG="$2"
        shift 2
        ;;
    -benchfunc)
        _BENCH_FUNC="$2"
        shift 2
        ;;
    -outputdir)
        _OUTPUTDIR="$2"
        shift 2
        ;;
    -template)
        _TEMPLATE="$2"
        shift 2
        ;;
    *)
        echo "Error: Invalid flag $1"
        usage
        exit 1
        ;;
    esac
done

# Check if all required flags were provided
if [[ -z $_PKG || -z $_BENCH_FUNC || -z $_OUTPUTDIR || -z $_TEMPLATE ]]; then
    echo "Error: Missing required flags."
    usage
    exit 1
fi

echo "ℹ️ Running benchmarks for package \"$_PKG\" and function \"$_BENCH_FUNC\"..."

mkdir -p "$_OUTPUTDIR"
_BENCHOUT=$(mktemp -p "$_OUTPUTDIR" "benchout.$_BENCH_FUNC.XXXXXX.txt")
go test -bench=^"$_BENCH_FUNC"$ -run=^$ -benchmem -benchtime=2s "$_PKG" |
    tee >(grep -v "^PASS\|^ok" >"$_BENCHOUT")

echo "Benchmarks saved at \"$_BENCHOUT\"."
echo "✔️ OK."

# format_output Formats the benchmark results into a markdown file
#
# Parameters:
#    input_file (str): The path to the file containing the benchmark results.
#    output_file (str): The path to the file where the formatted benchmark results will be saved.
#    template_file (str): The path to the template file used to format the benchmark results.
#    bencfunc (str): The benchmark function to run.
#
# Returns:
#    None
#
# Example:
#   format_output "outputdir.txt" "outputdir.md" "template.md"
format_output() {
    local input_file="$1"
    local output_file="$2"
    local template_file="$3"
    local bencfunc="$4"

    echo "ℹ️ Formatting benchmark results (input_file: $input_file, output_file: $output_file, template_file: $template_file)..."

    local benchmarks=""
    while IFS= read -r line; do
        if [[ $line == Benchmark* ]]; then
            local benchmark
            local ns_op
            local b_op
            local allocs_op
            benchmark=$(echo "$line" | awk '{print $1}')
            ns_op=$(echo "$line" | awk '{print $3}')
            b_op=$(echo "$line" | awk '{print $5}')
            allocs_op=$(echo "$line" | awk '{print $7}')
            benchmarks+="| $benchmark | $ns_op | $b_op | $allocs_op |\n"
        fi
    done <"$input_file"

    local runner=""
    if [[ "${CI:=false}" == "true" ]]; then
        local workflow_url="$GITHUB_SERVER_URL/$GITHUB_REPOSITORY/actions/runs/$GITHUB_RUN_ID"
        runner=$(echo -e "> The benchmark results were generated during a GitHub Actions workflow run on a $RUNNER_OS runner ([view workflow]($workflow_url)).")
    fi
    local goos
    local goarch
    local pkg
    local cpu
    local go_version
    local date
    local cmd
    goos=$(grep 'goos:' "$input_file" | awk -F': ' '{print $2}')
    goarch=$(grep 'goarch:' "$input_file" | awk -F': ' '{print $2}')
    pkg=$(grep 'pkg' "$input_file" | awk -F': ' '{print $2}')
    cpu=$(grep 'cpu' "$input_file" | awk -F': ' '{print $2}')
    go_version=$(go version | awk '{sub(/^go/, "", $3); print $3}')
    date=$(date +"%Y-%m-%d")
    cmd="go test -bench=^${bencfunc}$ -run=^$ -benchmem -benchtime=2s $pkg"
    sed -e "s#{{GOOS}}#$goos#g" \
        -e "s#{{GOARCH}}#$goarch#g" \
        -e "s#{{PKG}}#$pkg#g" \
        -e "s#{{CPU}}#$cpu#g" \
        -e "s#{{GO_VERSION}}#$go_version#g" \
        -e "s#{{DATE}}#$date#g" \
        -e "s#{{RUNNER_TEXT}}#$runner#g" \
        -e "s#{{CMD}}#$cmd#g" \
        -e "s#{{BENCHMARKS}}#$benchmarks#g" \
        "$template_file" >"$output_file"
    echo "✔️ OK. Formatted benchmark results saved in \"$output_file\"."
}

format_output "$_BENCHOUT" "$_OUTPUTDIR/benchmarks.md" "$_TEMPLATE" "$_BENCH_FUNC"

echo "✔️ Done."
