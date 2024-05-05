#!/bin/sh

for start_script in /etc/kickStart.d/*.sh
do
if [[ $start_script == "/etc/kickStart.d/health_check.sh" ]]; then
continue
fi
if [[ $start_script == "/etc/kickStart.d/dockerfile-entry.sh" ]]; then
continue
fi
echo "########################################## start exec $start_script"
sh $start_script
echo "########################################## done exec $start_script"
done

tail -f /dev/null


