#!/usr/bin/env bash
#
# Ensure the acctest omicron-dev docker image is available locally.
# Pulls from GHCR if available, otherwise builds locally.
#
# Usage:
#   ./ensure-image.sh <omicron-sha>

set -euo pipefail

OMICRON_SHA="${1:?usage: ensure-image.sh <omicron-sha>}"

REMOTE_IMAGE="ghcr.io/oxidecomputer/terraform-provider-oxide/acctest-omicron-dev:${OMICRON_SHA}"
LOCAL_IMAGE="acctest-omicron-dev:${OMICRON_SHA}"

if docker pull "${REMOTE_IMAGE}" 2>/dev/null; then
    docker tag "${REMOTE_IMAGE}" "${LOCAL_IMAGE}"
else
    echo "-> Image not found in GHCR, building locally"
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    docker compose --project-directory "${SCRIPT_DIR}" --file "${SCRIPT_DIR}/docker-compose.yaml" \
        build --build-arg "OMICRON_SHA=${OMICRON_SHA}"
fi
