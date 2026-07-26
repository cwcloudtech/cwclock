#!/usr/bin/env bash

set -e

source ./ci/cli/compute-env.sh
cd "${CWCLOCK_CLI_DIR}"

echo "Performing post-release cleanup..."
rm -rf dist || true
echo "✓ Cleanup completed"
