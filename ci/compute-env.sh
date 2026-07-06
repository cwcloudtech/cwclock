#!/usr/bin/env bash

export CWCLOUD_APPS="ui api"
export VERSION="${EDITION}-$(grep -oE "^[0-9\.]+$" VERSION)"
export VERSION_SHA="${VERSION}-${CI_COMMIT_SHORT_SHA}"
