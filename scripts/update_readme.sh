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
# Usage: ./update_readme.sh --dirpath [target_dir] --outfile [output_file]
#
# Example: ./update_readme.sh --dirpath ./path/to/directory --outfile ./path/to/output.md
#
# This script updates the README file by concatenating the individual README files in the provided directory.
# The following flags are required:
#
# --dirpath [string]       The directory containing the README files to concatenate.
#                          The files should follow the pattern:
#                              - 0001_filename.md
#                              - 0002_filename.md
#                              - 0003_filename.md
#                          where the number before the underscore describes the position of the file in the concatenation order.
#
# --outfile [string]       The target file to which the concatenated README files will be written.
#
# The script also creates a pull request with the updated README file.
#
# Note: This script is intended to be used in a CI/CD environment.
#-----------------------------------------------------------------------------------------------------------------

set -euo pipefail

_DIRPATH=""
_OUTFILE=""

usage() {
    echo "
Usage: $0 --dirpath [target_dir] --outfile [output_file]

Options:

    --dirpath [string]       The directory containing the README files to concatenate.
                             The files should follow the pattern:
                                 - 0001_filename.md
                                 - 0002_filename.md
                                 - 0003_filename.md
                             where the number before the underscore describes the position of the file in the concatenation order.

    --outfile [string]       The target file to which the concatenated README files will be written.
"
}

# Parse command line arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
    --dirpath)
        _DIRPATH="$2"
        shift
        ;;
    --outfile)
        _OUTFILE="$2"
        shift
        ;;
    *)
        echo "Unknown parameter passed: $1"
        usage
        exit 1
        ;;
    esac
    shift
done

# Validate required arguments
[[ -z "$_DIRPATH" ]] && echo "❌ --dirpath is required." && usage && exit 1
[[ ! -f "$_OUTFILE" ]] && echo "❌ --outfile is required." && usage && exit 1

# Get the sorted list of markdown files (absolute paths)
# Filters:
# - 'type f': only files
# - 'name 000*_*.md': files starting with '000' and ending with '.md'
# - 'size +0c': files with size greater than 0 bytes
# - 'perm /u=r': files with read permission for the owner
_MARKDOWN_FILES=$(find "$_DIRPATH" -type f -name '000*_*.md' -size +0c -perm /u=r -print | sort)
(($(echo "$_MARKDOWN_FILES" | wc -l) > 0)) || {
    echo "
❌ No markdown files found in $_DIRPATH

Ensure that the markdown files follow the pattern:
    - 0001_filename.md
    - 0002_filename.md
    - 0003_filename.md
where the number before the underscore describes the position of the file in the concatenation order.
"
    exit 1
}

# Get the URL of the script on GitHub
get_script_url() {
    # Get the name of the script
    script_name="$1"

    # Get the root of the git repository
    git_root=$(git rev-parse --show-toplevel)

    # Search for the path of the script in the git repository
    script_path=$(find "$git_root" -name "$script_name" -print -quit)

    # Get the relative path from the root of the git repository
    script_path=$(realpath --relative-to="$git_root" "$script_path")

    if [[ "${CI:=false}" == "true" ]]; then
        remote_url_https="$GITHUB_SERVER_URL/$GITHUB_REPOSITORY"
    else
        # Get the URL of the remote repository
        remote_url=$(git config --get remote.origin.url)

        # Convert SSH format to HTTPS format
        remote_url_https=$(echo "$remote_url" | sed -e 's/:/\//' -e 's/git@/https:\/\//')
    fi

    # Combine the remote URL and the script path
    echo "${remote_url_https%.git}/blob/master/${script_path}"
}

# Update the main README file
update_readme() {
    script_name="$1"
    script_url="$2"
    input_files="$3"
    outfile="$4"

    tmpfile=$(mktemp -p "$(dirname "$outfile")" "tmp.XXXXXXXXXX.md")
    trap 'rm -f "$tmpfile"' EXIT
    {
        for file in $input_files; do
            cat "$file"
            echo
            echo
        done
    } >"$tmpfile"
    # Append a note at the end of the README
    cat <<EOF >>"$tmpfile"
---

_Note: This file was auto-generated by the [$script_name](${script_url}) script. Do not edit this file directly._
EOF

    if [[ "${CI:=false}" == "true" ]]; then
        echo "ℹ️ Updating the README file..."
        mv "$tmpfile" "$outfile"
        echo "✔️ OK."
    else
        printf '%.0s-' {1..80}
        echo
        echo "ℹ️ Preview the updated README file:"
        cat "$tmpfile"
        echo
        printf '%.0s-' {1..80}
        echo
        echo "ℹ️ Diff of the updated README file:"
        diff -u "$outfile" "$tmpfile" || true
        printf '%.0s-' {1..80}
        echo
        echo "❓ Do you want to update the README file? (y/n)"
        read -r response
        if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
            mv "$tmpfile" "$outfile"
            echo "ℹ️ README file updated successfully."
        else
            echo "ℹ️ README file not updated."
        fi
    fi
}

# Create a pull request with the updated README
create_pull_request() {
    input_files="$1"
    outfile="$2"

    echo "ℹ️ Creating a pull request with the updated benchmark results..."

    # Switch back to the current branch on exit
    current_branch=$(git rev-parse --abbrev-ref HEAD)
    trap 'git checkout $current_branch' EXIT

    git config --global user.email "github-actions[bot]@users.noreply.github.com"
    git config --global user.name "github-actions[bot]"
    git config pull.rebase false
    # Add additional date to allow rerunning on a failed workflow (to avoid PR conflicts)
    _branch_name="automated-documentation-update-$GITHUB_RUN_ID-$(date +%s)"
    git checkout -b "$_branch_name"
    for file in $input_files; do
        git add "$file"
    done
    git add "$outfile"
    _commit_message="docs(README): Update benchmark results"
    git commit -m "${_commit_message} [skip ci]" || export NO_UPDATES=true
    if [[ "${NO_UPDATES:=false}" == "true" ]]; then
        echo "ℹ️ No updates to the _README file. Exiting..."
        exit 0
    fi
    git push origin "$_branch_name"
    gh pr create \
        --title "$_commit_message" \
        --body "Automated documentation update for benchmark results." \
        --label "documentation"

    echo "✔️ OK."
}

# Main function
main() {
    echo "ℹ️ Updating the main README file..."
    script_name=$(basename "$(realpath "$0")")
    script_url=$(get_script_url "$script_name")
    update_readme "$script_name" "$script_url" "$_MARKDOWN_FILES" "$_OUTFILE"
    if [[ "${CI:=false}" == "true" ]]; then
        create_pull_request "$_MARKDOWN_FILES" "$_OUTFILE"
    else
        echo "ℹ️ Not in a CI environment. Skipping pull request creation..."
    fi
    echo
    echo "🎉 README update complete!"
}
main "$@"
