#!/bin/bash

SWAYLOCK="swaylock -f -c 1f1626"

if pgrep -x swayidle > /dev/null; then
    pkill -x swayidle
    notify-send -t 2000 'Auto lock' 'DISABLED'
else
    swayidle -w \
        timeout 300 "${SWAYLOCK}" \
        timeout 600 'niri msg action power-off-monitors' \
        before-sleep "${SWAYLOCK}" \
        lock "${SWAYLOCK}" &
    notify-send -t 2000 'Auto lock' 'ENABLED'
fi
