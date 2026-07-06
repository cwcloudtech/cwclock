#!/usr/bin/env bash

export OS="$(uname -a)"
[[ $OS =~ .*Darwin.* ]] && export AWK_BIN="gawk" || export AWK_BIN="awk"

threshold=70
remaining_space=$(df -h|$AWK_BIN '($6 == "/"){match($5, /[0-9]+/, num); if (num[0] != ""){print num[0]} else {print "0"}}')

if [[ $remaining_space -ge $threshold ]]; then
    echo "[prune] executing docker system prune -af" 
    docker system prune -af
fi
