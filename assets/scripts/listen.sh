#!/usr/bin/env bash

################################################################
# Usage:                                                       #
#   ./listen.sh -s /tmp/cortile.sock                           #
#                                                              #
################################################################

usage() {
cat << EOF
Usage: $0 -s <sock file path> [-vh]

-s,  Sock file path of cortile process.
-v,  Run script in verbose mode.
-h,  Display help.

EOF
exit 1;
}

# Parse arguments
while getopts "s:vh" arg; do
    case ${arg} in
        s) sock=${OPTARG};;
        v) verbose=true;;
        h) usage;;
        *) usage;;
    esac
done

# Validate arguments
if [ $# -eq 0 ] || [ -z $sock ]; then
    usage
fi

# Check dependencies
for dep in nc jq; do
    [[ $(which $dep 2> /dev/null) ]] || { echo -e "'$dep' needs to be installed"; deps=1; }
done
[[ $deps -ne 1 ]] || { exit 1; }

# Define socket file
sockout="${sock}.out";
if [ $verbose ]; then
    echo -e "Socket: $sockout";
fi

# Listen to socket messages
while json=$(nc -Ulw 1 $sockout | jq -r "."); do
    if [ $verbose ]; then
        echo -e "\nMessage: $json";
    fi

    # Parse json data
    type=$(echo $json | jq -r ".Type")
    name=$(echo $json | jq -r ".Name")
    data=$(echo $json | jq -r ".Data")

    case ${type} in
        "Action")
            desk=$(echo $data | jq -r ".Desk")
            screen=$(echo $data | jq -r ".Screen")

            # EXAMPLE: retrieve action event on active workspace
            echo "Received 'action' with name '$name' on 'desktop = $desk' and 'screen = $screen'";;
        "State")
            case ${name} in
                "workspaces")
                    desk=$(echo $data | jq -r ".Desk")
                    screen=$(echo $data | jq -r ".Screen")                    
                    workspace=$(echo $data | jq -r ".Workspaces[] | select((.Location.DeskNum==$desk) and (.Location.ScreenNum==$screen))")
                    enabled=$(echo $workspace | jq -r ".TilingEnabled")
                    layout=$(echo $workspace | jq -r ".ActiveLayoutNum")

                    # EXAMPLE: retrieve tiling state and layout on active workspace
                    echo "Received '$name' with tiling 'enabled = $enabled' on 'desktop = $desk' and 'screen = $screen' with 'layout = $layout'";;
                "arguments")
                    config=$(echo $data | jq -r ".Config")

                    # EXAMPLE: retrieve config file path from command line arguments
                    echo "Received '$name' with name 'config = $config'";;
                "configs")
                    decoration=$(echo $data | jq -r ".WindowDecoration")

                    # EXAMPLE: retrieve window decoration setting from config file
                    echo "Received '$name' with property 'WindowDecoration = $decoration'";;
            esac
    esac
done
