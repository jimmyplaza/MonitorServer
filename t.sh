#!/bin/sh

sh checkMonitor.sh
if [ $? -eq '0' ]; then
   echo "START MONITORING..."
fi
