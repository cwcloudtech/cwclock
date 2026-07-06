#!/usr/bin/env bash

source ./ci/compute-env.sh

echo "" > .env.cwcloud.db
env|grep "POSTGRES_"|while read; do
  echo "${REPLY}" >> .env.cwcloud.db
done

echo "" > .env.cwcloud.api
echo "API_URL=${API_URL}" > .env.cwcloud.ui

docker ps -a | grep -i cwclock | awk '{system ("docker rm -f "$1)}'
docker compose -f docker-compose-live.yml up -d --force-recreate
