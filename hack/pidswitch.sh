#!/bin/sh

# podname=$1
echo "hello10"
# # SLEEPPID=$(for pid in $(pgrep sleep); do if grep --silent ${podname} /proc/$pid/environ; then echo $pid;fi; done)
# SLEEPPID=000
# # for pid in $(pgrep sleep); do if grep --silent ${podname} /proc/$pid/environ; then $SLEEPPID=$pid;fi; done
# for pid in $(pgrep sleep); do
#     echo $pid;
#     if grep --silent ${podname} /proc/$pid/environ;
#     then  SLEEPPID=$pid;
#           fi;
# done;
# echo "pid: ${SLEEPPID}"

# echo "done"


SLEEPPID=1 # because we are not running in the host pid namespace
/tmp/squash-client  > /proc/$SLEEPPID/fd/1 2> /proc/$SLEEPPID/fd/2

