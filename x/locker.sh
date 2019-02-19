#!/bin/bash

export XSECURELOCK_PARANOID_PASSWORD=0

if command -v /usr/lib/xscreensaver/flyingtoasters >/dev/null; then
    export XSECURELOCK_SAVER=/usr/lib/xscreensaver/flyingtoasters
fi

export XSECURELOCK_FONT="Iosevka-21"
export XSECURELOCK_DISCARD_FIRST_KEYPRESS=1

for cmd in /usr/share/goobuntu-desktop-files/xsecurelock.sh xsecurelock i3lock; do
    if command -v $cmd >/dev/null; then
        exec $cmd
    fi
done
