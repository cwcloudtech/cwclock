#!/usr/bin/env bash

source ./ci/compute-env.sh

sha="$(git rev-parse --short HEAD)"
details="$(git log --pretty=format:"%an, %ar : %s" -1)"

echo '{"version":"'"${VERSION}"'", "env":"'"${ENV}"'", "sha":"'"${sha}"'", "details":"'"${details}"'"}' > manifest.json

docker login "${CI_REGISTRY}" --username "${CI_REGISTRY_USER}" --password "${CI_REGISTRY_PASSWORD}"

for app in "${CWCLOCK_APPS}"; do
  export IMAGE_NAME="cwclock-${app}"
  export SERVICE_NAME="${app}"

  docker buildx bake --push ${SERVICE_NAME}
  if [[ $VERSION != $VERSION_SHA ]]; then
    docker tag "${CI_REGISTRY}/${IMAGE_NAME}:${VERSION}" "${CI_REGISTRY}/${IMAGE_NAME}:${VERSION_SHA}"
    docker push "${CI_REGISTRY}/${IMAGE_NAME}:${VERSION_SHA}"
  fi
done
