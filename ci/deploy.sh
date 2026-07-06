#!/usr/bin/env bash

source ./ci/compute-env.sh

echo "" > .env.cwclock.db
env|grep "POSTGRES_"|while read; do
  echo "${REPLY}" >> .env.cwclock.db
done

echo "" > .env.cwclock.api
echo "API_URL=${API_URL}" > .env.cwclock.ui

docker ps -a | grep -i cwclock | awk '{system ("docker rm -f "$1)}' || :
docker compose -f docker-compose-live.yml up -d --force-recreate
docker logs cwclock-db-migrate || :
