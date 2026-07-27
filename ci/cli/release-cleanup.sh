#!/usr/bin/env bash

set -e

source ./ci/cli/compute-env.sh
cd "${CWCLOCK_CLI_DIR}"

echo "Cleaning up any existing release artifacts..."
if [ -d "dist" ]; then
  rm -rf dist
  echo "✓ Cleaned up dist directory"
fi
