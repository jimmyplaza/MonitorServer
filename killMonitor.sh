#!/bin/sh

gomanage=`ps -ef|grep MonitorServer|grep -v 'grep' |awk {'print $2'}`

for i in $gomanage; do
	kill $i
done;
