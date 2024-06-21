#!/usr/bin/env bash
set -euo pipefail

COVERPROFILE="${COVERPROFILE:-cover.out}"
WORKSPACE="${GITHUB_WORKSPACE:-$(git rev-parse --show-toplevel)}"
COVERDIR="${COVERDIR:-$WORKSPACE/.cover}"
mkdir -p "$COVERDIR"
gomoddirs=(
    '.'
    './drivers/postgres'
    './middleware/echo'
    './middleware/nethttp'
)

echo "Running tests..."
for dir in "${gomoddirs[@]}"; do
    [[ -f "$dir/go.mod" ]] || {
        echo "No go.mod file found at $dir"
        exit 1
    }
    printf '\n\n%s\n' "$(printf '=%.0s' {1..80})"
    printf "🐛 Testing module at path: %s\n" "$dir"
    printf '%s\n' "$(printf '=%.0s' {1..80})"
    if [[ "$(basename "$dir")" == "." ]]; then
        coverfile="$COVERDIR/root.out"
    else
        coverfile="$COVERDIR/$(basename "$dir").out"
    fi
    pushd "$dir" >/dev/null
    go test -v -race -outputdir="$COVERDIR" -coverprofile="$coverfile" -timeout 15m ./...
    popd >/dev/null
done

# Upload coverage report
printf '\n\n%s\n' "$(printf '=%.0s' {1..80})"
echo "Uploading coverage report..."
printf '%s\n' "$(printf '=%.0s' {1..80})"
if [[ "${CI:-false}" == "true" ]]; then
    bash <(curl -s https://codecov.io/bash) -f "$COVERDIR"'/*.out' -p "$WORKSPACE"
else
    echo "Skipping upload. Not running in CI."
fi
rm -rf "$COVERDIR"

echo -e "\nDone."
