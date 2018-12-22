#!/bin/sh
consul connect proxy -sidecar-for socat-rpi-0 &
consul connect proxy -sidecar-for monitor-rpi-0 &
