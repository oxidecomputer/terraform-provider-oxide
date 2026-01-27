#!/usr/bin/env bash
#
# Output the omicron commit SHA to use for acceptance tests.
#
# Usage:
#   ./omicron-version.sh [VERSION]
#
# If VERSION is provided (branch name or SHA), resolve it to a full SHA.
# Otherwise, look up the version from oxide.go's VERSION_OMICRON file.

set -euo pipefail

OMICRON_VERSION="${1:-}"

if [[ -z "${OMICRON_VERSION}" ]]; then
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

    OXIDE_COMMIT=$(cd "${REPO_ROOT}" && go list -m -json github.com/oxidecomputer/oxide.go | jq -r '.Version | split("-") | .[-1]')
    if [[ -z "${OXIDE_COMMIT}" ]]; then
        echo "Error: Could not extract oxide.go commit from go.mod" >&2
        exit 1
    fi

    OMICRON_VERSION=$(curl -sL "https://raw.githubusercontent.com/oxidecomputer/oxide.go/${OXIDE_COMMIT}/VERSION_OMICRON")
    if [[ -z "${OMICRON_VERSION}" ]]; then
        echo "Error: Could not fetch VERSION_OMICRON from oxide.go" >&2
        exit 1
    fi
fi

gh api "repos/oxidecomputer/omicron/commits/${OMICRON_VERSION}" --jq '.sha'
