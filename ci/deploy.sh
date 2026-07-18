#!/usr/bin/env bash

source ./ci/compute-env.sh

echo "" > .env.cwclock.db
env|grep "POSTGRES_"|while read; do
  echo "${REPLY}" >> .env.cwclock.db
done

echo "" > .env.cwclock.api
env|grep -E "(CWCLOCK|CWCLOUD)_"|while read; do
  echo "${REPLY}" >> .env.cwclock.api
done

echo "API_URL=${API_URL}" > .env.cwclock.ui
echo "CWCLOCK_MAX_IMAGE_SIZE=${CWCLOCK_MAX_IMAGE_SIZE}" >> .env.cwclock.ui

docker ps -a | grep -i cwclock | awk '{system ("docker rm -f "$1)}' || :
docker compose -f docker-compose-live.yml up -d --force-recreate
docker logs cwclock-db-migrate || :
