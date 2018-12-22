#!/bin/sh
rm -rf /var/consul/*
HOSTNAME=`hostname -I | awk '{print $1}'`
consul agent \
	-config-dir=./consul.d \
	-bind=${HOSTNAME} \
	-enable-script-checks=true \
	-http-port=8500 \
	-client=127.0.0.1 \
	-log-file="/tmp/consul.log" &
