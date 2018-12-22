#!/bin/sh
consul connect proxy -sidecar-for socat-rpi-1 &
consul connect proxy -sidecar-for monitor-rpi-1 &
