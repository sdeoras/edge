#!/bin/sh
consul connect proxy -sidecar-for socat-rpi-2 &
consul connect proxy -sidecar-for monitor-rpi-2 &
