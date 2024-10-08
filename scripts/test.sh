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
    ["./examples"]="${gotestflagsbase[@]} -coverpkg=./..."
)

cleanup() {
    # shellcheck disable=SC2317
    unset workspace coverprofile modcoverage gotestflagsbase gotestflags coverageflags _temp
}
trap cleanup EXIT

for dir in "${!gotestflags[@]}"; do
    echo "================================================================================"
    echo "🧪 Testing module at path: $dir"
    echo "================================================================================"
    pushd "$dir" >/dev/null
    # shellcheck disable=SC2086
    go test ${gotestflags[$dir]} ./...
    cat "$modcoverage" >>"$coverprofile"
    go tool cover -html="$modcoverage" -o "${modcoverage%.out}.html"
    popd >/dev/null
done

echo "✅ All tests passed!"

[[ ${UPLOAD_COVERAGE:-false} != "true" ]] && exit 0

for dir in "${!gotestflags[@]}"; do
    echo "================================================================================"
    echo "📊 Generating coverage report using codecov for module at path: $dir"
    echo "================================================================================"
    modflag=$(basename "$workspace")
    if [[ "$dir" != "." ]]; then
        modflag+="/${dir#./}"
    fi
    # for more opts see: bash <(curl -s https://codecov.io/bash) -help
    coverageflags=(
        '-f' "${dir}/$modcoverage"
        '-p' "${workspace}${dir#.}"
        '-F' "${modflag//\//_}"
    )
    [[ ${CI:-false} != "true" ]] && coverageflags+=('-d')

    bash <(curl -s https://codecov.io/bash) "${coverageflags[@]}"
    echo "================================================================================"
    echo "🎉 Coverage report generated for for module at path: $dir"
    echo "================================================================================"
done

echo "✅ Done. All coverage reports uploaded to Codecov!"
