#!/bin/bash

SCRIPT_DIR="$(realpath "$(dirname "$BASH_SOURCE")")"

# For older versions of xsecurelock
export XSECURELOCK_PARANOID_PASSWORD=0

# For newer versions of xsecurelock
export XSECURELOCK_PASSWORD_PROMPT=asterisks

screensavers=(
    /usr/lib/xscreensaver/flyingtoasters
    /usr/libexec/xscreensaver/flyingtoasters
)

for saver in "${screensavers[@]}"; do
    if command -v "${saver}" >/dev/null; then
        export XSECURELOCK_SAVER="${saver}"
        break
    fi
done

export XSECURELOCK_FONT="Iosevka-21"
export XSECURELOCK_DISCARD_FIRST_KEYPRESS=0

for cmd in xsecurelock i3lock; do
    if command -v $cmd >/dev/null; then
        exec $cmd
    fi
done
