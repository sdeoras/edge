#!/bin/sh
rm -rf /var/consul/*
HOSTNAME=`hostname -I | awk '{print $1}'`
consul agent \
	-config-dir=./consul.d \
	-dev \
	-ui \
	-bind=${HOSTNAME} \
	-client="0.0.0.0" \
	-enable-script-checks=true \
	-log-file="/tmp/consul.log" &
