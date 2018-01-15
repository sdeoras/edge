#!/bin/bash
set -ex
GOOS=linux GOARCH=arm GOARM=7 go build -o ${GOPATH}/bin/snapshot-server .
