#!/usr/bin/env bash
set -euo pipefail

printf "=%.0s" {1..80}
echo "🔧 Updating dependencies..."
find . \
    -name 'go.mod' \
    -exec printf '=%.0s' {1..80} \; \
    -printf '\ngo.mod directory: %h\n\n' \
    -printf '\n:: 🛠️ Updating...\n' \
    -execdir go get -u \; \
    -printf '\n  ✔️ OK. Updated.' \
    -printf '\n::🔧 Running go mod tidy...\n' \
    -execdir go mod tidy \; \
    -printf '\n  ✔️ OK. Tidied.'
    
echo
echo "✅ Done. All dependencies updated."