#!/bin/bash

LOCKER="${HOME}/.local/bin/swaylock-toasters"

if pgrep -x swayidle > /dev/null; then
    pkill -x swayidle
    notify-send -t 2000 'Auto lock' 'DISABLED'
else
    swayidle -w \
        timeout 300 "${LOCKER}" \
        timeout 600 'swaymsg "output * dpms off"' \
        resume 'swaymsg "output * dpms on"' \
        before-sleep "${LOCKER}" \
        lock "${LOCKER}" &
    notify-send -t 2000 'Auto lock' 'ENABLED'
fi
