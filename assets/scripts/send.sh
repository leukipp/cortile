#!/usr/bin/env bash

################################################################
# Usage:                                                       #
#   ./send.sh -s /tmp/cortile.sock -m '{"Action":"tile"}'      #
#   ./send.sh -s /tmp/cortile.sock -m '{"Action":"untile"}'    #
#   ./send.sh -s /tmp/cortile.sock -m '{"Action":"........"}'  #
#   ./send.sh -s /tmp/cortile.sock -m '{"State":"workspaces"}' #
#   ./send.sh -s /tmp/cortile.sock -m '{"State":"arguments"}'  #
#   ./send.sh -s /tmp/cortile.sock -m '{"State":"configs"}'    #
#                                                              #
################################################################

usage() {
cat << EOF
Usage: $0 -s <sock file path> -m '{"<type>":"<command>"}' [-vh]

-s,  Sock file path of cortile process.
-m,  Message string to be transmitted.
-v,  Run script in verbose mode.
-h,  Display help.

EOF
exit 1;
}

# Parse arguments
while getopts "s:m:vh" arg; do
    case ${arg} in
        s) sock=${OPTARG};;
        m) msg=${OPTARG};;
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
sockin="${sock}.in";
if [ $verbose ]; then
    echo -e "Socket: $sockin";
fi

# Send socket message
if [ $msg ]; then
    if [ $verbose ]; then
        echo -e "Message: $msg";
    fi
    echo $msg | nc -Uw 1 $sockin
fi
