#!/usr/bin/env bash

# Set up dependencies for the acceptance test suite. The tests expect various
# resources to be configured.

set -euo pipefail

PROJECT_NAME=${OXIDE_PROJECT:-tf-acc-test}

# Default to test-suite-silo, the silo used by omicron-dev.
SILO_NAME=${OXIDE_SILO:-test-suite-silo}

# Build a sample image, if not specified by caller.
IMAGE_PATH=${OXIDE_IMAGE_PATH:-alpine.raw}
if [ ! -e "$IMAGE_PATH" ]; then
    curl -L -o alpine.qcow2 https://dl-cdn.alpinelinux.org/alpine/v3.22/releases/cloud/generic_alpine-3.22.1-x86_64-bios-tiny-r0.qcow2
    qemu-img convert -f qcow2 -O raw alpine.qcow2 "$IMAGE_PATH"
fi

if ! oxide project view --project $PROJECT_NAME > /dev/null; then
    oxide project create --name $PROJECT_NAME --description $PROJECT_NAME
fi

# We need to create disks, images, etc., so override the default empty quota.
oxide silo quotas update --silo $SILO_NAME --cpus 100 --memory $((2 ** 40)) --storage $((2 ** 40))

# Set up the default IP pool, and add a range.
if ! oxide ip-pool view --pool default > /dev/null; then
    oxide ip-pool create --name default --description default
    oxide ip-pool silo link --pool default --silo $SILO_NAME --is-default true
    oxide ip-pool range add --first 10.0.1.0 --last 10.0.1.255 --pool default
fi

# The acceptance tests expect both at least a single project-scoped image and a
# silo-scoped image. Import the same image twice, then promote one copy to the
# silo. Use alpine because it's small.
oxide disk import \
    --project $PROJECT_NAME \
    --path $IMAGE_PATH \
    --disk alpine-project \
    --description "alpine image" \
    --snapshot alpine-snapshot-project \
    --image alpine-project \
    --image-description "alpine image" \
    --image-os alpine \
    --image-version "3.22.1"

oxide image promote --image alpine-project --project $PROJECT_NAME

oxide disk import \
    --project $PROJECT_NAME \
    --path $IMAGE_PATH \
    --disk alpine-silo \
    --description "alpine image" \
    --snapshot alpine-snapshot-silo \
    --image alpine-silo \
    --image-description "alpine image" \
    --image-os alpine \
    --image-version "3.22.1"
