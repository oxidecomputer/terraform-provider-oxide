#!/usr/bin/env bash

# Upsert dependencies for testing configuration files in examples/. Assumes
# acc-test-setup.sh has already run.

set -euo pipefail

if ! oxide project view --project my-project > /dev/null; then
    oxide project create --name my-project --description my-project
fi

if ! oxide instance anti-affinity view --project my-project --anti-affinity-group my-group > /dev/null 2>&1; then
    oxide instance anti-affinity create \
        --project my-project \
        --name my-group \
        --description my-group \
        --policy allow \
        --failure-domain sled > /dev/null
fi

if ! oxide disk view --project my-project --disk my-disk > /dev/null 2>&1; then
    oxide api "/v1/disks?project=my-project" --method POST --input - > /dev/null <<'EOF'
{
  "name": "my-disk",
  "description": "my-disk",
  "size": 1073741824,
  "disk_backend": {
    "type": "distributed",
    "disk_source": { "type": "blank", "block_size": 512 }
  }
}
EOF
fi

if ! oxide current-user ssh-key view --ssh-key my-key > /dev/null 2>&1; then
    oxide current-user ssh-key create \
        --name my-key \
        --description my-key \
        --public-key "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAINV+93cl3H+Nk6LDnkiTJ7MAiUsvC34qb/gFN2DZ1pej my-key@my-project" > /dev/null
fi
