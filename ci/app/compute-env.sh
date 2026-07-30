#!/usr/bin/env bash

export APPS="ui api ui-and-mobile"
export APP_PREFIX="cwclock"
export VERSION="$(grep -oE "^[0-9\.]+$" VERSION)"
export VERSION_SHA="${VERSION}-${CI_COMMIT_SHORT_SHA}"
export CWCLOCK_API_URL="https://api.cwclock.me"
export CWCLOCK_UI_URL="https://www.cwclock.me"
export CWCLOCK_CORS_ENABLED="off"
export CWCLOCK_MAX_IMAGE_SIZE="2097152"
export CWCLOCK_OTEL_PROTO="otlp/grpc"
export CWCLOCK_MAX_REPORT_SIZE=5000
export CWCLOCK_ACTIVATION_MODE="email"
export CWCLOCK_LIMIT_MAIL=100

echo "VERSION=${VERSION}" > .env.ci
echo "VERSION_SHA=${VERSION_SHA}" >> .env.ci
