#!/bin/bash
consul connect proxy -service socat-client-rpi-0 -upstream socat-rpi-0:9190 &
consul connect proxy -service socat-client-rpi-1 -upstream socat-rpi-0:9191 &
consul connect proxy -service socat-client-rpi-2 -upstream socat-rpi-0:9192 &

consul connect proxy -service monitor-client-rpi-0 -upstream monitor-rpi-0:60050 &
consul connect proxy -service monitor-client-rpi-1 -upstream monitor-rpi-1:60051 &
consul connect proxy -service monitor-client-rpi-2 -upstream monitor-rpi-2:60052 &
