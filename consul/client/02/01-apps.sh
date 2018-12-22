#!/bin/sh
socat -v tcp-l:8181,fork exec:"/bin/cat" &
monitor-server -host=127.0.0.1 -tag=garage &
