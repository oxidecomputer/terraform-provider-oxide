#!/usr/bin/env bash

# Get the relevant commit of oxide.rs that best corresponds to a given omicron version. First look
# up the OpenAPI spec version corresponding to the omicron ref. Then find the most recent commit to
# oxide.rs@main that uses the expected spec version. If no match is found, use the latest commit to
# oxide.rs@main.

set -euo pipefail

OMICRON_VERSION="${1:?Usage: oxide-cli-version.sh OMICRON_VERSION}"

echo "Omicron version: ${OMICRON_VERSION}" >&2

# Look up the openapi spec version from the omicron version.
NEXUS_LATEST=$(curl -sL "https://raw.githubusercontent.com/oxidecomputer/omicron/${OMICRON_VERSION}/openapi/nexus/nexus-latest.json")
if [[ -z "${NEXUS_LATEST}" ]]; then
    echo "Error: Could not fetch nexus-latest.json" >&2
    exit 1
fi
SPEC_URL="https://raw.githubusercontent.com/oxidecomputer/omicron/${OMICRON_VERSION}/openapi/nexus/${NEXUS_LATEST}"
API_VERSION=$(curl -sL "${SPEC_URL}" | jq -r '.info.version')

if [[ -z "${API_VERSION}" || "${API_VERSION}" == "null" ]]; then
    echo "Error: Could not extract API version from ${SPEC_URL}" >&2
    exit 1
fi
echo "Nexus API version: ${API_VERSION}" >&2

# Search oxide.rs commits for matching version.
COMMITS=$(curl -sL "https://api.github.com/repos/oxidecomputer/oxide.rs/commits?path=oxide.json&per_page=50" | jq -r '.[].sha')
for sha in ${COMMITS}; do
    version=$(curl -sL "https://raw.githubusercontent.com/oxidecomputer/oxide.rs/${sha}/oxide.json" | jq -r '.info.version' 2>/dev/null || echo "")
    if [[ "${version}" == "${API_VERSION}" ]]; then
        echo "Found oxide.rs commit ${sha} with API version ${version}" >&2
        echo "${sha}"
        exit 0
    fi
done

# If no exact match, default to `main`. We're probably using omicron main in
# this case, so the latest oxide.rs is a reasonable default.
MAIN_SHA=$(git ls-remote https://github.com/oxidecomputer/oxide.rs refs/heads/main | cut -f1)
echo "No exact match for API version ${API_VERSION}. Defaulting to oxide.rs main (${MAIN_SHA})." >&2
echo "${MAIN_SHA}"
