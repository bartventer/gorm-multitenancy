#!/usr/bin/env bash
set -euo pipefail

# This script runs tests and generates coverage reports for all modules in the project.

workspace="${GITHUB_WORKSPACE:-$(git rev-parse --show-toplevel)}"
coverprofile="${COVERPROFILE:-$workspace/cover.out}"

modcoverage="modcoverage.out"
gotestflagsbase=(
    '-v'
    '-race'
    '-timeout' '20m'
    "-coverprofile=$modcoverage"
)

declare -A gotestflags
gotestflags=(
    ["."]="${gotestflagsbase[@]} -coverpkg=./..."
    ["./postgres"]="${gotestflagsbase[@]} -coverpkg=./..."
    ["./mysql"]="${gotestflagsbase[@]} -coverpkg=./..."
    ["./middleware/echo"]="${gotestflagsbase[@]}"
    ["./middleware/gin"]="${gotestflagsbase[@]}"
    ["./middleware/iris"]="${gotestflagsbase[@]}"
    ["./middleware/nethttp"]="${gotestflagsbase[@]}"
    ["./_examples"]="${gotestflagsbase[@]} -coverpkg=./..."
)

cleanup() {
    # shellcheck disable=SC2317
    unset workspace coverprofile modcoverage gotestflagsbase gotestflags coverageflags _temp
}
trap cleanup EXIT

run_tests() {
    local dir="$1"
    echo "================================================================================"
    echo "🧪 Testing module at path: $dir"
    echo "================================================================================"
    # shellcheck disable=SC2086
    go test ${gotestflags[$dir]} ./... 2>&1 | tee "test.log"
    [[ ${CI:-false} != "true" ]] && tail -n +2 "$modcoverage" >>"$coverprofile"
}

if [[ ${CI:-false} != "true" ]]; then
    mkdir -p "$(dirname "$coverprofile")"
    echo "mode: atomic" >"$coverprofile"
fi

while read -r dir; do
    pushd "$dir" >/dev/null
    run_tests "$dir"
    popd >/dev/null
done < <(printf "%s\n" "${!gotestflags[@]}" | sort)

[[ ${CI:-false} != "true" ]] && go tool cover -html="$coverprofile" -o "${coverprofile%.out}.html"

echo "✅ All tests passed!"

[[ ${UPLOAD_COVERAGE:-false} != "true" ]] && exit 0

echo "================================================================================"
echo "📊 Uploading combined coverage report to Codecov"
echo "================================================================================"

coverageflags=(
    '-F' "combined"
    '-X' "coveragepy"
    '-X' "recursesubs"
)

if [[ ${CI:-false} != "true" ]]; then
    coverageflags+=('-d')
    bash <(curl -s https://codecov.io/bash) "${coverageflags[@]}" >"$workspace/codecov-upload.out"
    echo "Upload file written to $workspace/codecov-upload.out"
else
    bash <(curl -s https://codecov.io/bash) "${coverageflags[@]}"
fi

echo "================================================================================"
echo "🎉 Combined coverage report uploaded to Codecov"
echo "================================================================================"
