#!/bin/bash

if xset q | grep '  timeout:  0 ' >/dev/null; then
    xset s 300 5
    notify-send --expire-time=2000 'Auto lock ENABLED!'
else
    xset s off
    notify-send --expire-time=2000 'Auto lock DISABLED!'
fi
