#!/usr/bin/env bash
#
# Build the acctest omicron-dev Docker image.
#
# Checks for a cached image in the registry first. If found, pulls and tags it
# locally. Otherwise, builds the image from scratch.
#
# Usage:
#   ./build.sh [OPTIONS] [OMICRON_VERSION]
#
# Arguments:
#   OMICRON_VERSION  - Branch name or commit SHA. If not provided, looks up
#                      the version from oxide.go's VERSION_OMICRON file.
#
# Options:
#   --force          - Build even if image exists in cache.
#
# Environment variables:
#   IMAGE_REGISTRY   - Registry for cache lookup (default: ghcr.io)
#   IMAGE_REPO       - Registry repository path (default: oxidecomputer/terraform-provider-oxide)
#
# Examples:
#   ./build.sh                    # Use cached image or build with version from oxide.go
#   ./build.sh main               # Use cached image or build with omicron main
#   ./build.sh --force            # Force rebuild ignoring cache
#   ./build.sh --force abc123     # Force rebuild with specific commit

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Parse arguments
FORCE_BUILD=false
OMICRON_VERSION=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --force)
            FORCE_BUILD=true
            shift
            ;;
        -*)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
        *)
            OMICRON_VERSION="$1"
            shift
            ;;
    esac
done

# Look up omicron version from oxide.go if not provided
if [[ -z "${OMICRON_VERSION}" ]]; then
    echo "Looking up omicron version from oxide.go..."
    OMICRON_VERSION=$("${SCRIPT_DIR}/omicron-version.sh")
    echo "  Found omicron version: ${OMICRON_VERSION}"
fi

IMAGE_REGISTRY="${IMAGE_REGISTRY:-ghcr.io}"
IMAGE_REPO="${IMAGE_REPO:-oxidecomputer/terraform-provider-oxide}"
IMAGE_NAME="acctest-omicron-dev"
# Sanitize version for use as a tag (replace non-alphanumeric with _)
IMAGE_TAG="$(echo "${OMICRON_VERSION}" | sed 's/[^[:alnum:]]/_/g')"

LOCAL_IMAGE="${IMAGE_NAME}:${IMAGE_TAG}"
REMOTE_IMAGE="${IMAGE_REGISTRY}/${IMAGE_REPO}/${IMAGE_NAME}:${IMAGE_TAG}"

echo "Omicron version: ${OMICRON_VERSION}"
echo "Local image: ${LOCAL_IMAGE}"
echo "Remote image: ${REMOTE_IMAGE}"

# Output variables for GitHub Actions
if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
    echo "local_image=${LOCAL_IMAGE}" >> "${GITHUB_OUTPUT}"
    echo "remote_image=${REMOTE_IMAGE}" >> "${GITHUB_OUTPUT}"
fi

# Check cache unless --force is specified
if [[ "${FORCE_BUILD}" != "true" ]]; then
    echo "Checking for cached image..."
    if docker pull "${REMOTE_IMAGE}" 2>/dev/null; then
        echo "Using cached image: ${REMOTE_IMAGE}"
        docker tag "${REMOTE_IMAGE}" "${LOCAL_IMAGE}"
        echo "Tagged as: ${LOCAL_IMAGE}"
        exit 0
    fi
    echo "Cached image not found, building..."
fi

echo "Building image..."
docker build \
    --build-arg "OMICRON_BRANCH=${OMICRON_VERSION}" \
    --tag "${LOCAL_IMAGE}" \
    --tag "${REMOTE_IMAGE}" \
    "${SCRIPT_DIR}"

echo "Built: ${LOCAL_IMAGE}"
echo "Tagged as: ${REMOTE_IMAGE}"
