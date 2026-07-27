#!/usr/bin/env bash

export CWCLOCK_CLI_DIR="./cwclock-cli"
export VERSION="${CI_COMMIT_TAG}"
export VERSION_SHA="${CI_COMMIT_TAG}"

echo "VERSION=${VERSION}" > .env.ci
echo "VERSION_SHA=${VERSION_SHA}" >> .env.ci
