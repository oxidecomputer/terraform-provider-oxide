#!/usr/bin/env bash
#
# Output the omicron version from oxide.go's VERSION_OMICRON file.
#
# Usage:
#   ./omicron-version.sh
#
# The version is determined by:
# 1. Finding the oxide.go commit from go.mod
# 2. Fetching VERSION_OMICRON from oxide.go at that commit

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Extract oxide.go commit from go.mod (format: v0.x.y-date-commit)
OXIDE_COMMIT=$(grep 'oxide.go' "${REPO_ROOT}/go.mod" | sed 's/.*-\([a-f0-9]*\)$/\1/')
if [[ -z "${OXIDE_COMMIT}" ]]; then
    echo "Error: Could not extract oxide.go commit from go.mod" >&2
    exit 1
fi

# Fetch VERSION_OMICRON from oxide.go at that commit
OMICRON_VERSION=$(curl -sL "https://raw.githubusercontent.com/oxidecomputer/oxide.go/${OXIDE_COMMIT}/VERSION_OMICRON")
if [[ -z "${OMICRON_VERSION}" ]]; then
    echo "Error: Could not fetch VERSION_OMICRON from oxide.go" >&2
    exit 1
fi

echo "${OMICRON_VERSION}"
