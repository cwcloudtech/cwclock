#!/usr/bin/env bash

declare -A CWCLOUD_APPS
export CWCLOUD_APPS=("ui" "api")

export CI_REGISTRY="rg.fr-par.scw.cloud/cwclock-t8d7th"

export VERSION="${EDITION}-$(grep -oE "^[0-9\.]+$" VERSION)"
export VERSION_SHA="${VERSION}-${CI_COMMIT_SHORT_SHA}"
