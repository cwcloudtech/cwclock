#!/usr/bin/env bash

app="${1}"
env_file=".env.ci"

[ -f "${env_file}" ] && source "${env_file}"

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

discord_notif "${DISCORD_TOKEN_PUBLIC}" "cwclock_${app}"
