#!/usr/bin/env bash

source ./ci/app/compute-env.sh

sha="$(git rev-parse --short HEAD)"
details="$(git log --pretty=format:"%an, %ar : %s" -1)"

echo '{"version":"'"${VERSION}"'", "sha":"'"${sha}"'", "details":"'"${details}"'"}' > manifest.json

docker login "${CI_REGISTRY}" --username "${CI_REGISTRY_USER}" --password "${CI_REGISTRY_PASSWORD}"

for app in $CWCLOCK_APPS; do
  image_name="cwclock-${app}"

  local_version="${VERSION}"
  local_version_sha="${VERSION_SHA}"
  local_latest="latest"

  if [[ $app = "ui-and-mobile" ]]; then
    local_version="${VERSION}-mobile"
    local_version_sha="${VERSION_SHA}-mobile"
    local_latest="latest-mobile"
    image_name="ui"
  fi

  docker buildx bake -f docker-compose-build.yml --push "${app}"
  if [[ $VERSION != $VERSION_SHA ]]; then
    docker tag "${CI_REGISTRY}/${image_name}:${local_version}" "${CI_REGISTRY}/${image_name}:${local_version_sha}"
    docker push "${CI_REGISTRY}/${image_name}:${local_version_sha}"
  fi

  docker tag "${CI_REGISTRY}/${image_name}:${local_version}" "${CI_REGISTRY}/${image_name}:${local_latest}"
  docker push "${CI_REGISTRY}/${image_name}:${local_version}"
  docker push "${CI_REGISTRY}/${image_name}:${local_latest}"
done
