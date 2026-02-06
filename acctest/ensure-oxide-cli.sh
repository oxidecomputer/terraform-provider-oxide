#!/usr/bin/env bash
#
# Ensure the correct version of the oxide CLI is installed for the given
# omicron version. Binaries are cached by oxide.rs commit SHA in ./bin/oxide-<sha>,
# with ./bin/oxide symlinked to the current version.
#
# Usage:
#   ./ensure-oxide-cli.sh [omicron-sha]
#
# If omicron-sha is not provided, fetch it from go.mod.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
BIN_DIR="${REPO_ROOT}/bin"

OMICRON_SHA="${1:-$(${SCRIPT_DIR}/omicron-version.sh)}"
OXIDE_RS_COMMIT=$("${SCRIPT_DIR}/oxide-cli-version.sh" "${OMICRON_SHA}")
CLI_PATH="${BIN_DIR}/oxide-${OXIDE_RS_COMMIT}"

if [[ ! -x "${CLI_PATH}" ]]; then
    echo "Installing oxide CLI from oxide.rs@${OXIDE_RS_COMMIT}"
    cargo install \
        --git https://github.com/oxidecomputer/oxide.rs \
        --rev "${OXIDE_RS_COMMIT}" \
        --root "${REPO_ROOT}" \
        oxide-cli
    mv "${BIN_DIR}/oxide" "${CLI_PATH}"
fi

ln -sf "oxide-${OXIDE_RS_COMMIT}" "${BIN_DIR}/oxide"
echo "oxide CLI: ${OXIDE_RS_COMMIT}"
