#!/bin/bash

# Useful for quickly checking out a HSL stop's timetable

# add your preferred stops here in the following format:
# "Nickname:Stop name:Stop code" 
# Stop code can be found on the HSL reittiopas or in hslterm by searching by name:
#    hslterm -stop=STOP
#       This will give you a list of stops with the name STOP and their codes

STOPS=(
    "🏠 Vuosaari metroasema:Vuosaari:H0040"
    "🏫 Aalto-yliopisto metroasema (itään menevät):Aalto:E0003"
)

TERM="x-terminal-emulator -e"

# function to print list of stops in format index Nickname (Stop name, Stop Code) 
function print_stops {
    for stop in "${STOPS[@]}"; do
        IFS=':' read -r -a stop_info <<< "$stop"
        echo "${stop_info[0]} (${stop_info[1]}, ${stop_info[2]})"
    done
    echo "cancel"
}

SELECTION=$(print_stops | rofi -dmenu -i -p "Select stop (or press esc to search)")

# if cancel is selected, exit
if [ "$SELECTION" == "cancel" ]; then
    exit 0
fi

# separate selection string by string separated by character '('
_='' read -r -a selection_info <<< "$SELECTION"

# get length of list selection_info
len=${#selection_info[@]}

STOPNAME=$(echo ${selection_info[$len-2]:1:-1})
STOPCODE=$(echo ${selection_info[$len-1]::-1})

echo $STOPNAME
echo $STOPCODE

$TERM "hslterm -stop=$STOPNAME -code=$STOPCODE -tui"
