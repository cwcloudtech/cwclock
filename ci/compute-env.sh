#!/usr/bin/env bash

export CWCLOCK_APPS="ui api"
export VERSION="$(grep -oE "^[0-9\.]+$" VERSION)"
export VERSION_SHA="${VERSION}-${CI_COMMIT_SHORT_SHA}"
export API_URL="https://api.cwclock.me"
export CWCLOCK_CORS_ENABLED="off"
export IMAGE_SIZE="256"
