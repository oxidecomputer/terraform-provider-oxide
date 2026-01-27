#!/usr/bin/env bash
#
# Find the oxide.rs commit that matches a given omicron version.
#
# Usage:
#   ./oxide-cli-version.sh [OMICRON_VERSION]
#
# If OMICRON_VERSION is not provided, looks it up from oxide.go's VERSION_OMICRON.
#
# The script:
# 1. Gets the nexus API version from the omicron commit
# 2. Searches oxide.rs commits to find one with matching API version
# 3. Outputs the oxide.rs commit SHA

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

OMICRON_VERSION="${1:-}"

# Look up omicron version if not provided
if [[ -z "${OMICRON_VERSION}" ]]; then
    OMICRON_VERSION=$("${SCRIPT_DIR}/omicron-version.sh")
fi

echo "Omicron version: ${OMICRON_VERSION}" >&2

# Get the nexus API version from omicron
# The file nexus-latest.json contains just the filename of the latest spec
NEXUS_LATEST=$(curl -sL "https://raw.githubusercontent.com/oxidecomputer/omicron/${OMICRON_VERSION}/openapi/nexus/nexus-latest.json")
# Extract version from filename like "nexus-2026010800.0.0-1844ae.json"
API_VERSION=$(echo "${NEXUS_LATEST}" | sed -n 's/nexus-\([0-9.]*\)-.*/\1/p')

if [[ -z "${API_VERSION}" ]]; then
    echo "Error: Could not extract API version from nexus-latest.json" >&2
    exit 1
fi

echo "Nexus API version: ${API_VERSION}" >&2

# Search oxide.rs commits for matching version
# Get recent commits that modified oxide.json (newest first)
COMMITS=$(curl -sL "https://api.github.com/repos/oxidecomputer/oxide.rs/commits?path=oxide.json&per_page=50" | jq -r '.[].sha')

BEST_MATCH=""
BEST_VERSION=""

for sha in ${COMMITS}; do
    version=$(curl -sL "https://raw.githubusercontent.com/oxidecomputer/oxide.rs/${sha}/oxide.json" | jq -r '.info.version' 2>/dev/null || echo "")

    # Exact match - use it
    if [[ "${version}" == "${API_VERSION}" ]]; then
        echo "Found exact match: oxide.rs ${sha} (API ${version})" >&2
        echo "${sha}"
        exit 0
    fi

    # Track the newest version that's not newer than what we need
    # Version format is YYYYMMDDXX.X.X, so string comparison works
    if [[ -n "${version}" && "${version}" < "${API_VERSION}" ]]; then
        if [[ -z "${BEST_MATCH}" || "${version}" > "${BEST_VERSION}" ]]; then
            BEST_MATCH="${sha}"
            BEST_VERSION="${version}"
        fi
    fi
done

# Use best match if we found one
if [[ -n "${BEST_MATCH}" ]]; then
    echo "Warning: No exact match found. Using newest compatible oxide.rs commit." >&2
    echo "  Omicron API: ${API_VERSION}" >&2
    echo "  Oxide.rs API: ${BEST_VERSION} (${BEST_MATCH})" >&2
    echo "${BEST_MATCH}"
    exit 0
fi

echo "Error: Could not find oxide.rs commit with API version <= ${API_VERSION}" >&2
exit 1
