#!/bin/bash

LOCKER="/home/jrosenth/.local/bin/swaylock-toasters"

if pgrep -x swayidle > /dev/null; then
    pkill -x swayidle
    notify-send -t 2000 'Auto lock' 'DISABLED'
else
    swayidle -w \
        timeout 300 "${LOCKER}" \
        timeout 600 'niri msg action power-off-monitors' \
        before-sleep "${LOCKER}" \
        lock "${LOCKER}" &
    notify-send -t 2000 'Auto lock' 'ENABLED'
fi
