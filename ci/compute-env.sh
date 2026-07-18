#!/usr/bin/env bash

export CWCLOCK_APPS="ui api"
export VERSION="$(grep -oE "^[0-9\.]+$" VERSION)"
export VERSION_SHA="${VERSION}-${CI_COMMIT_SHORT_SHA}"
export API_URL="https://api.cwclock.me"
export CWCLOCK_UI_URL="https://www.cwclock.me"
export CWCLOCK_CORS_ENABLED="off"
export CWCLOCK_MAX_IMAGE_SIZE="2097152"
export CWCLOCK_OTEL_PROTO="otlp/grpc"
export CWCLOCK_MAX_REPORT_SIZE=5000
export CWCLOCK_ACTIVATION_MODE="email"
