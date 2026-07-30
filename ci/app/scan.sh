#!/usr/bin/env bash

source ./ci/app/compute-env.sh

for app in $APPS; do
  image_name="${APP_PREFIX}-${app}"
  local_version="${VERSION}"

  if [[ $app = "ui-and-mobile" ]]; then
    local_version="${VERSION}-mobile"
    image_name="${APP_PREFIX}-ui"
  fi

  echo "Pulling recent built image"
  docker login "${CI_REGISTRY}" --username "${CI_REGISTRY_USER}" --password "${CI_REGISTRY_PASSWORD}"
  docker pull "${CI_REGISTRY}/${image_name}:${local_version}"

  echo "Scanning recent build image"
  trivy image \
  --exit-code ${CI_SCAN_EXIT_CODE} \
  --severity ${CI_SCAN_SEVERITY} "${CI_REGISTRY}/${image_name}:${local_version}"
done
