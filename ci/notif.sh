#!/usr/bin/env bash

source ./ci/compute-env.sh

discord_notif() {
    token="${1}"
    username="${2}"
    if [[ $token ]]; then
        message="${username} has been successfully deployed! version = ${VERSION}"
        endpoint="https://discord.com/api/webhooks/${token}/slack"
        payload="{\"text\": \"${message}\", \"username\": \"${username}\"}"
        curl -X POST "${endpoint}" -H "Accept: application/json" -d "${payload}"
    fi
}


for app in "${CWCLOCK_APPS[@]}"; do
  discord_notif "${DISCORD_TOKEN_PUBLIC}" "cwclock_${CWCLOCK_APP}"
done
