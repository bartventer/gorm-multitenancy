#!/usr/bin/env bash
set -euxo pipefail

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
        coverfile="$COVERDIR/rover.cover"
    else
        coverfile="$COVERDIR/$(basename "$dir").cover"
    fi
    pushd "$dir" >/dev/null
    go test -v -race -outputdir="$COVERDIR" -coverprofile="$coverfile" -timeout 15m ./...
    popd >/dev/null
done

# Merge all coverage files
echo "Merging coverage files..."
echo "mode: atomic" >"$COVERPROFILE"
tail -q -n +2 "$COVERDIR"/*.cover >>"$COVERPROFILE"
rm -rf "$COVERDIR"

echo "Done."
