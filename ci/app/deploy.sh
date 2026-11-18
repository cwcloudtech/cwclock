#!/usr/bin/env bash

source ./ci/app/compute-env.sh

echo "" > .env.cwclock.db
# Anchor to a real "VAR=" assignment line and read -r, so a multi-line value
# (e.g. a cert/token with embedded newlines) can't have its continuation lines
# matched on a bare substring and split into bogus env entries.
env|grep -E "^POSTGRES_[A-Z0-9_]*="|while read -r line; do
  echo "${line}" >> .env.cwclock.db
done

echo "JWT_SECRET=${JWT_SECRET}" > .env.cwclock.api
# DISCORD_TOKEN isn't CWCLOCK_/CWCLOUD_-prefixed, so (like JWT_SECRET) it's written
# explicitly here, from the runner's DISCORD_TOKEN_PUBLIC (mapped in compute-env.sh).
echo "DISCORD_TOKEN=${DISCORD_TOKEN}" >> .env.cwclock.api
env|grep -E "^(CWCLOCK|CWCLOUD)_[A-Z0-9_]*="|while read -r line; do
  echo "${line}" >> .env.cwclock.api
done

echo "CWCLOCK_API_URL=${CWCLOCK_API_URL}" > .env.cwclock.ui
echo "CWCLOCK_UI_URL=${CWCLOCK_UI_URL}" >> .env.cwclock.ui
echo "CWCLOCK_MAX_IMAGE_SIZE=${CWCLOCK_MAX_IMAGE_SIZE}" >> .env.cwclock.ui

docker ps -a | grep -i cwclock | awk '{system ("docker rm -f "$1)}' || :
docker login "${CI_REGISTRY}" --username "${CI_REGISTRY_USER}" --password "${CI_REGISTRY_PASSWORD}"
if ! docker buildx imagetools inspect "${CI_REGISTRY}/cwclock-ui:${UI_VERSION}" > /dev/null 2>&1; then
  export UI_VERSION="${VERSION}"
  echo "Deploying non-mobile version"
fi
docker compose -f docker-compose-live.yml up -d --force-recreate
docker logs cwclock-db-migrate || :
