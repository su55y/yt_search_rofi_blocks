#!/bin/sh

SCRIPTPATH="$(cd -- "$(dirname "$0")" >/dev/null 2>&1 || exit 1 ; pwd -P)"
[ ! -f "$SCRIPTPATH/yt_search_rofi_blocks" ] && \
    notify-send "yt_search_rofi_blocks" "blocks wrapper not found"

# C-h -- prev page
# C-l -- next page
# C-c -- clear

rofi  -modi blocks \
    -show blocks \
    -theme yt_search.rasi \
    -normal-window \
    -blocks-wrap "$SCRIPTPATH/yt_search_rofi_blocks" \
    -kb-mode-complete "Control+Alt+l" \
    -kb-remove-char-back "BackSpace,Shift+BackSpace" \
    -kb-custom-1 "Control+h" \
    -kb-custom-2 "Control+l" \
    -kb-custom-3 "Control+c"
