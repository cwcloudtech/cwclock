#!/usr/bin/env bash

source ./ci/app/compute-env.sh

echo "" > .env.cwclock.db
env|grep "POSTGRES_"|while read; do
  echo "${REPLY}" >> .env.cwclock.db
done

echo "JWT_SECRET=${JWT_SECRET}" > .env.cwclock.api
env|grep -E "(CWCLOCK|CWCLOUD)_"|while read; do
  echo "${REPLY}" >> .env.cwclock.api
done

echo "CWCLOCK_API_URL=${CWCLOCK_API_URL}" > .env.cwclock.ui
echo "CWCLOCK_UI_URL=${CWCLOCK_UI_URL}" >> .env.cwclock.ui
echo "CWCLOCK_MAX_IMAGE_SIZE=${CWCLOCK_MAX_IMAGE_SIZE}" >> .env.cwclock.ui

docker ps -a | grep -i cwclock | awk '{system ("docker rm -f "$1)}' || :
docker compose -f docker-compose-live.yml up -d --force-recreate
if [[ $? != 0 ]]; then
  export UI_VERSION="${VERSION}"
  echo "Deploying non-mobile version"
  docker compose -f docker-compose-live.yml up -d --force-recreate
fi
docker logs cwclock-db-migrate || :
