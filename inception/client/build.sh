#!/bin/bash
set -ex
GOOS=linux GOARCH=arm GOARM=6 go build -o ${GOPATH}/bin/inception-client .
